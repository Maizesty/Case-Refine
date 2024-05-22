
# from sanic.signals import register_signals
import yaml
import numpy as np
from rpc_scheduler import rpc_factory
from rpc_proxy import rpc_proxy
from fastapi import FastAPI, HTTPException
import numpy as np
from pydantic import BaseModel
import uvicorn
import cachetools
app = FastAPI()

    

with open('./conf/config.yaml', 'r') as file:
    config = yaml.safe_load(file)
lb_setting = config['location'][0]
lb_model = lb_setting['balance_mode']
workers = lb_setting['proxy_pass']
cache_size = config['cache_size']
scheduler = rpc_factory.get_proxy(lb_model,workers,cache_size=cache_size,warm_up_times=config["warm_up_times"])
proxy = rpc_proxy.RPCProxy(workers,scheduler)
class Item(BaseModel):
    Keys: list

# @app.post("/predict")
# async def predict(request):
#     json_data = request.json
#     keys = json_data.get('Keys', [])
#     np_array = np.array(keys).astype(np.int64)
#     result = proxy.iner(np_array)
#     return response.json({
#         'data': result.tolist()
#     })
@app.post("/predict")
async def predict(item:Item):
    # return {"data": 1}
    keys = item.Keys
    if not keys:
        raise HTTPException(status_code=400, detail="Keys field is required")
    try:
        np_array = np.array(keys).astype(np.int64)
        result =  await proxy.iner(np_array)
        return {"data": result.tolist()}
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

if __name__ == '__main__':

    uvicorn.run("server:app", host="0.0.0.0", port=config['port'], reload=True, log_level="critical", workers=1)
    # app.run(host='0.0.0.0', port=config['port'])