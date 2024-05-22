FROM nvcr.io/nvidia/tritonserver:24.04-py3
RUN ln -s /usr/lib/x86_64-linux-gnu/libcudnn.so.9.1.0 /usr/lib/x86_64-linux-gnu/libcudnn.so.8
RUN pip3 install cachetools tritonclient[all] redis -i  https://pypi.tuna.tsinghua.edu.cn/simple
RUN pip3 install onnxruntime-gpu --extra-index-url https://aiinfra.pkgs.visualstudio.com/PublicPackages/_packaging/onnxruntime-cuda-12/pypi/simple/