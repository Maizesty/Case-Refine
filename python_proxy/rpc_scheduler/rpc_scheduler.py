from abc import ABC, abstractmethod

# 定义一个抽象基类
class RPCSchdeuler(ABC):

    # 定义一个抽象方法
    @abstractmethod
    def add(self,host:str):
      pass
    
    @abstractmethod
    def remove(self,host:str):
      pass
    
    @abstractmethod
    def schedule(self,reuest):
      pass
    
    @abstractmethod
    def inc(self,host:str):
      pass
    
    @abstractmethod
    def done(self,host:str):
      pass