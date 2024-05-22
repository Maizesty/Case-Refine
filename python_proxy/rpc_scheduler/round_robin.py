from rpc_scheduler.base_rpc_scheduler import BaseRPCScheduler

from atomic import AtomicLong
import threading



class RoundRobin(BaseRPCScheduler):
  def __init__(self, workers) -> None:
    # self.i = 0
    self.i = 0
    self._lock = threading.RLock()
    super().__init__(workers)
  def inc(self, host: str):
    pass
  def schedule(self, reuest):
    # with self._lock:
    #   index = self.i % len(self.workers)
    #   self.i +=1
    # with self._lock:
    value = self.i
    self.i += 1
    index = value % len(self.workers)
    return self.workers[index]
    
