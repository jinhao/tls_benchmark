# tls_benchmark

## Desc

a https session cache test tool.

## Clone repositity

> git clone git@github.com:jinhao/tls_benchmark.git  
> git submodule init  
> git submodule update  

## 注意

由于测试中使用了虚拟ip，因此跨机器测试时，需要在服务端添加路由信息,如在91.101上添加虚拟ip172.16.30.1-254, 则在服务端按如下命令添加路由 

`route add -net 172.16.30.0  netmask 255.255.255.0 gw 172.16.91.101`
