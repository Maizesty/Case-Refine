import json
import threading
import time
import cachetools
import numpy
import tritonclient.grpc as grpcclient
import triton_python_backend_utils as pb_utils
from tritonclient.utils import *
import os
import redis

class TritonPythonModel:
    def initialize(self, args):
        """`initialize` is called only once when the model is being loaded.
        Implementing `initialize` function is optional. This function allows
        the model to initialize any state associated with this model.

        Parameters
        ----------
        args : dict
          Both keys and values are strings. The dictionary keys and values are:
          * model_config: A JSON string containing the model configuration
          * model_instance_kind: A string containing model instance kind
          * model_instance_device_id: A string containing model instance device ID
          * model_repository: Model repository path
          * model_version: Model version
          * model_name: Model name
        """

        # You must parse model_config. JSON string is not parsed here
        self.model_config = model_config = json.loads(args["model_config"])
        cache_size = eval(model_config['parameters']['cache_size']['string_value'])
        self.caches = cachetools.LRUCache(cache_size)
 
        # Get OUT configuration
        out_config = pb_utils.get_output_config_by_name(model_config, "OUT")

        self.triton_client = grpcclient.InferenceServerClient(url="49.52.27.77:18001")
        # Convert Triton types to numpy types
        self.out_dtype = pb_utils.triton_string_to_numpy(out_config["data_type"])
        self.redis_client = redis.Redis(host='49.52.27.77', port=6379, db=0)
        self.hit_metric_family = pb_utils.MetricFamily(
            name="hit",
            description="hit",
            kind=pb_utils.MetricFamily.COUNTER,  # or pb_utils.MetricFamily.GAUGE
        )

        # Create a Metric object under the MetricFamily object. The 'labels'
        # is a dictionary of key-value pairs. You can create multiple Metric
        # objects under the same MetricFamily object with unique labels. Empty
        # labels is allowed. The 'labels' parameter is optional. If you don't
        # specify the 'labels' parameter, empty labels will be used.
        self.hit_metric = self.hit_metric_family.Metric(
            labels={"model": "embed_cache", "version": "1"}
        )
        self.total_metric_family = pb_utils.MetricFamily(
            name="total",
            description="total",
            kind=pb_utils.MetricFamily.COUNTER,  # or pb_utils.MetricFamily.GAUGE
        )
        self.total_metric = self.total_metric_family.Metric(
            labels={"model": "embed_cache", "version": "1"}
        )
        # To keep track of response threads so that we can delay
        # the finalizing the model until all response threads
        # have completed.
        self.inflight_thread_count = 0
        self.inflight_thread_count_lck = threading.Lock()
        
        
    def execute(self, requests):
        responses = []
        start = time.time()
        for request in requests:
            responses.append(self.process_request(request))
        end = time.time()
        # print(f"cache time {(end - start) * 1000},len(requests):{len(requests)}")
        return responses

    def response_thread(self, in_input):
        # The response_sender is used to send response(s) associated with the
        # corresponding request.  Iterate over input/delay pairs. Wait for DELAY
        # milliseconds and then create and send a response.

        out_dtype = self.out_dtype
        start_init = time.time()
        # num_request = in_input.size//26
        # result = numpy.empty((in_input.size//26,26,16), dtype=out_dtype)
        result = numpy.random.rand(in_input.size//26,26,16).astype(out_dtype)
        
        need_key_dict = {}
        cnt = 0
        for i,arr in enumerate(in_input):
            for j, id in enumerate(arr):
                if id not in self.caches:
                    if id not in need_key_dict:
                        need_key_dict[id] = [(i,j)]
                    else :
                        
                        need_key_dict[id].append((i,j))
                else:
                    cnt+=1
                    result[i][j] = self.caches[id]
        # fake the embedding that missed
        need_key = [i for i in need_key_dict.keys()]
        end_init = time.time()
        # print(f"time of init{(end_init - start_init) * 1000}")
        if(len(need_key) == 0) :
            # print("perfect")
            pass
        else:
            start = time.time()
            # embed_arr = self.embed_remote_call(need_key)
            embed_arr = self.embed_redis_call(need_key)
            for index , key in enumerate(need_key):
                embed = embed_arr[index]
                self.caches[key] = embed
                for (i,j) in need_key_dict[key]:
                    result[i][j] = embed
            end = time.time()
            # print(f"time of remote{(end - start) * 1000}")
        # start_concat = time.time()
        # ll = [self.caches[j] for i  in in_input for j in  i  ]
        # result = numpy.array(ll).reshape(in_input.size//26,26,16).astype(out_dtype)
        # end_concat = time.time()
        # print(f"time of concat{(end_concat - start_concat) * 1000}")
        out_output = pb_utils.Tensor("OUT", result)
        response = pb_utils.InferenceResponse(
            output_tensors=[out_output]
        )
        self.hit_metric.increment(cnt)
        self.total_metric.increment(len(in_input))
        return response
        # response_sender.send(response)
        # # print("当前进程的PID是：", os.getpid())

        # # We must close the response sender to indicate to Triton that we are
        # # done sending responses for the corresponding request. We can't use the
        # # response sender after closing it. The response sender is closed by
        # # setting the TRITONSERVER_RESPONSE_COMPLETE_FINAL.
        # response_sender.send(flags=pb_utils.TRITONSERVER_RESPONSE_COMPLETE_FINAL)

        # with self.inflight_thread_count_lck:
        #     self.inflight_thread_count -= 1
    def process_request(self, request):
        # Start a separate thread to send the responses for the request. The
        # sending back the responses is delegated to this thread.
        # thread = threading.Thread(
        #     target=self.response_thread,
        #     args=(
        #         request.get_response_sender(),
        #         pb_utils.get_input_tensor_by_name(request, "IN").as_numpy(),
        #     ),
        # )
        return self.response_thread( 
                pb_utils.get_input_tensor_by_name(request, "IN").as_numpy())
        # A model using decoupled transaction policy is not required to send all
        # responses for the current request before returning from the execute.
        # To demonstrate the flexibility of the decoupled API, we are running
        # response thread entirely independent of the execute thread.
        # thread.daemon = True

        # with self.inflight_thread_count_lck:
        #     self.inflight_thread_count += 1

        # thread.start()
        
    def embed_redis_call(self,need_key):
        keys = [f"key_{i}" for i in need_key]
        arrays_bytes =  self.redis_client.mget(keys)
        # for k, v in enumerate(arrays_bytes):
        #     if v is None:
        #         print(f"misskey : {keys[k]}")
        arrays = [numpy.frombuffer(arr, dtype=numpy.float32) if arr else numpy.random.rand(1,16).astype(numpy.float32) for arr in arrays_bytes ]
        return arrays
        
    def embed_remote_call(self,need_key):
        model_name = "hps_embedding"
        need_key_len = len(need_key)
        input0_data = numpy.array([need_key]).astype(numpy.int64)
        input1_data = numpy.full((1, 1),need_key_len).astype(numpy.int32)
        inputs = [
        grpcclient.InferInput(
            "KEYS", input0_data.shape, np_to_triton_dtype(input0_data.dtype)
        ),
        grpcclient.InferInput(
            "NUMKEYS", input1_data.shape, np_to_triton_dtype(input1_data.dtype)
        ),
        ]

        inputs[0].set_data_from_numpy(input0_data)
        inputs[1].set_data_from_numpy(input1_data)

        outputs = [
        grpcclient.InferRequestedOutput("OUTPUT0"),
        ]
        response = self.triton_client.infer(model_name, inputs, request_id=str(0), outputs=outputs)

        output0_data = response.as_numpy("OUTPUT0")
        output0_data = output0_data.reshape(-1,16)
        # output0_data = numpy.random.random((need_key_len,26,16)).astype(self.out_dtype)
        return output0_data
    def finalize(self):
        """`finalize` is called only once when the model is being unloaded.
        Implementing `finalize` function is OPTIONAL. This function allows
        the model to perform any necessary clean ups before exit.
        Here we will wait for all response threads to complete sending
        responses.
        """
        print("Finalize invoked")

        inflight_threads = True
        cycles = 0
        logging_time_sec = 5
        sleep_time_sec = 0.1
        cycle_to_log = logging_time_sec / sleep_time_sec
        while inflight_threads:
            with self.inflight_thread_count_lck:
                inflight_threads = self.inflight_thread_count != 0
                if cycles % cycle_to_log == 0:
                    print(
                        f"Waiting for {self.inflight_thread_count} response threads to complete..."
                    )
            if inflight_threads:
                time.sleep(sleep_time_sec)
                cycles += 1

        print("Finalize complete...")