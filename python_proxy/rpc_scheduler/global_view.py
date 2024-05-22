import cachetools
from rpc_scheduler.base_rpc_scheduler import BaseRPCScheduler
from atomic import AtomicLong
import random


class GlobalView(BaseRPCScheduler):
  def __init__(self, workers,cache_size,warm_up_times) -> None:
    self.caches = {host:cachetools.LRUCache(cache_size) for host in workers}
    self.count = 0
    self.warm_up_times = warm_up_times 
    super().__init__(workers)
  def inc(self, host: str):
    pass  
  def schedule(self, requests):
    value = self.count
    self.count +=1
    if self.count < self.warm_up_times:
      index = value % len(self.workers)
      w = self.workers[index]
      cache = self.caches[w]
      for request in requests:
        for key in request:
          cache[key] = key
      return w
    else : 
      index = []
      best_score = 0
      for i in self.workers:
        cache = self.caches[i]
        score = 0
        for request in requests:
          for key  in request:
            if key in cache:
              score +=1
        if score > best_score:
          # print(f"change to {i}")
          index = [i]
          best_score = score
        elif score == best_score:
          index.append(i)
      if len(index) != 0:
        i = random.choice(index)
        cache = self.caches[i]
        for request in requests:
          for key in request:
            cache[key] = key
        # print(len(index))
        return i
      else:
        index = value % self.worker_num
        return self.workers[index]