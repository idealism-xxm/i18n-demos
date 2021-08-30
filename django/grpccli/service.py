import grpc

from grpccli import interceptor
from grpccli.gen.gin_pb2 import HelloRequest
from grpccli.gen.gin_pb2_grpc import GinServiceStub


def hello(name: str) -> str:
    with grpc.insecure_channel('host.docker.internal:50051') as channel:
        # 对 channel 应用所有拦截器
        channel = interceptor.intercept_channel_all(channel)
        stub = GinServiceStub(channel)
        # 如果不想使用拦截器，也可以在调用的时候指定带 trace 信息的 metadata
        response = stub.Hello(HelloRequest(name=name))

    return response.message
