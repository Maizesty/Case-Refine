import pickle as pl
import random
from rpc_scheduler.base_rpc_scheduler import BaseRPCScheduler

class CommiunityAware(BaseRPCScheduler):
  def __init__(self, workers) -> None:
    with open('/home/yssun/proxy_partition/partition.pkl','rb') as f:
      self.partition = pl.load(f)
      self.hot_set = pl.load(f)
    self.cnts = [0 for _ in range(4)]
    super().__init__(workers)
  def inc(self, host: str):
    pass  
  def genIndex(self,index):
    cnt = self.cnts[index]
    cnt_index = cnt % 3
    self.cnts[index]  = self.cnts[index] + 1
    return self.workers[index * 3 + cnt_index]

  def schedule(self, requests):
    maxVal = 0
    maxIndex = 0
    cnt = [0 for _ in range(4)]
    for request in requests:
      for key  in request:
        if key % 73880486 in self.hot_set:
          continue
        else:
          index = self.partition[key% 73880486]
          cnt[index] = cnt[index] + 1
          if cnt[index] > maxVal:
            maxVal = cnt[index]
            maxIndex = index
     
          
    return self.genIndex(maxIndex)


