
import tritonclient.grpc.aio as grpcclient
import rpc_scheduler
import rpc_scheduler.round_robin
import random 
from tritonclient.utils import *
import time
class RPCProxy:
  def __init__(self,hosts,scheduler) -> None:
    self.hostMap = {host:grpcclient.InferenceServerClient(host) for host in hosts}
    # todo
    self.scheduler = scheduler
  
  
  async def iner(self,request):
    start_schedule = time.time()
    index = self.scheduler.schedule(request)
    end_schedule = time.time()
    # print(f"schedule_time:{(end_schedule - start_schedule) * 1000}")
    start_prepare = time.time()
    client = self.hostMap[index]
    inputs = [
        grpcclient.InferInput(
            "IN", request.shape, np_to_triton_dtype(request.dtype)
        )
    ]
    inputs[0].set_data_from_numpy(request)
    outputs = [
        grpcclient.InferRequestedOutput("output"),
    ]
    end_prepare = time.time()
    # print(f"prepare_time:{(end_prepare - start_prepare) * 1000}")
    start_infer = time.time()
    response = await client.infer("sync_query", inputs, request_id=str(random.randint(1,10000000)), outputs=outputs)
    end_infer  = time.time()
    # print(f"infer_time:{(end_infer - start_infer) * 1000}")
    output0_data = response.as_numpy("output")
    return output0_data