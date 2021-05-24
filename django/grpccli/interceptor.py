from elasticapm.conf import constants
from elasticapm.traces import execution_context
from grpc import UnaryUnaryClientInterceptor
from grpc._interceptor import _ClientCallDetails


class TraceUnaryUnaryClientInterceptor(UnaryUnaryClientInterceptor):
    def intercept_unary_unary(self, continuation, client_call_details, request):
        # 先获取原有的 metadata
        metadata = []
        if client_call_details.metadata is not None:
            metadata = list(client_call_details.metadata)

        # 加上 trace 的相关信息
        transaction = execution_context.get_transaction()
        if transaction and transaction.trace_parent:
            value = transaction.trace_parent.to_string()
            metadata.append((constants.TRACEPARENT_HEADER_NAME, value))

        # 生成最新的 details
        new_details = _ClientCallDetails(
            client_call_details.method,
            client_call_details.timeout,
            metadata,
            client_call_details.credentials,
            client_call_details.wait_for_ready,
            client_call_details.compression,
        )

        return continuation(new_details, request)


trace_unary_unary_client_interceptor = TraceUnaryUnaryClientInterceptor()
