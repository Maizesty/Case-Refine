from rpc_scheduler.round_robin import RoundRobin
from rpc_scheduler.global_view import GlobalView
from rpc_scheduler.commiunity_aware import CommiunityAware
def get_proxy(balance_mode,worker, **kwargs):
  print(kwargs)
  if balance_mode=="round-robin":
    return RoundRobin(worker)
  if balance_mode=="global_view":
    return GlobalView(workers=worker,cache_size=kwargs['cache_size'],warm_up_times=kwargs['warm_up_times'])
  if balance_mode=="ca":
    return CommiunityAware(workers=worker)