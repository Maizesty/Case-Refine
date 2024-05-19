各个部分的用处
# dev-ops
用于存储构建自定义Triton Server的Dockerfile和启动服务的docker-compose

# gen_embed
用于生成存储在HPS上Embedding文件
# hps_demo
用于存储hps模型
# models
存储用于构建infer服务的模型
- sync_query：整个服务的入口，通过使用自定义逻辑实现查cache，推理，返回结果
- embed_cache：embedding cache 服务
- onnx_model_dfm/onnx_model：用于推理使用的模型，分别是deepFM和FM模型，（注意：当前版本将onnx模型删除了，后续会补上）