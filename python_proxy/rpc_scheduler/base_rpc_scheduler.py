from rpc_scheduler.rpc_scheduler import RPCSchdeuler

import threading
class Smap:
    def __init__(self,workers):
        self._lock = threading.RLock()
        self.M = {i:0 for i in workers}

    def get(self, key):
        with self._lock:
            return self.M.get(key, None)

    def set(self, key, value):
        with self._lock:
            self.M[key] = value
            
    def inc(self,key):
        with self._lock:
          self.M[key] += 1

    def delete(self, key):
        with self._lock:
            if key in self.M:
                del self.M[key]

    def update(self, key, value):
        with self._lock:
            self.M[key] = self.M.get(key, 0) + value

class BaseRPCScheduler(RPCSchdeuler):
  def __init__(self,workers) -> None:
    super().__init__()
    self.workers = workers
    self.lock = threading.RLock()
    self.cnt = Smap(workers)
    
  def add(self, host: str):
      with self.lock:
        if host not in self.workers:
          self.workers.append(host)
      return super().add(host)
  
  def remove(self, host: str):
      with self.lock:
        self.workers = [i for i in self.workers if i != host]
      return super().remove(host)
  
  def done(self, host: str):
    pass
    # self.cnt.inc(host)
    # return super().done(host)
    