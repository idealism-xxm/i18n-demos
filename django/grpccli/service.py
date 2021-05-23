import grpc

from grpccli.gen.gin_pb2 import HelloRequest
from grpccli.gen.gin_pb2 import HelloResponse
from grpccli.gen.gin_pb2_grpc import GinServiceStub
from grpccli.interceptor import trace_unary_unary_client_interceptor

def hello(name: str) -> str:
    with grpc.insecure_channel('host.docker.internal:50051') as channel:
        channel = grpc.intercept_channel(channel, trace_unary_unary_client_interceptor)
        stub = GinServiceStub(channel)
        response = stub.Hello(HelloRequest(name=name))

    return response.message
