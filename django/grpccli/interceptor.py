import grpc
from typing import List
from typing import Tuple

from django.utils import timezone
from django.utils.translation import trans_real
from elasticapm.conf import constants
from elasticapm.traces import execution_context
from grpc import UnaryUnaryClientInterceptor
from grpc._interceptor import _ClientCallDetails


GRPC_HEADER_NAME_TIMEZONE = 'x-timezone'
GRPC_HEADER_NAME_LANGUAGE = 'x-language'


class CustomUnaryUnaryClientInterceptor(UnaryUnaryClientInterceptor):
    """
    自定义一元客户端拦截器
    使用模版方法模式，简化子类加入 metadata 信息
    """
    def intercept_unary_unary(self, continuation, client_call_details, request):
        # 先获取原有的 metadata
        metadata = []
        if client_call_details.metadata is not None:
            metadata = list(client_call_details.metadata)

        # 获取子类中得到的 metadata 信息列表，放入到 metadata 中
        metadata.extend(
            self._get_metadata_list(continuation, client_call_details, request)
        )

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

    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        """获取 metadata 信息列表（子类重写）"""
        return []


class TraceUnaryUnaryClientInterceptor(CustomUnaryUnaryClientInterceptor):
    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        # 返回 trace 的相关信息
        transaction = execution_context.get_transaction()
        if transaction and transaction.trace_parent:
            value = transaction.trace_parent.to_string()
            return [(constants.TRACEPARENT_HEADER_NAME, value)]

        return []


class LanguageUnaryUnaryClientInterceptor(CustomUnaryUnaryClientInterceptor):
    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        # 返回 语言 的相关信息
        return [(GRPC_HEADER_NAME_LANGUAGE, trans_real.get_language())]


class TimezoneUnaryUnaryClientInterceptor(CustomUnaryUnaryClientInterceptor):
    def _get_metadata_list(self, continuation, client_call_details, request) -> List[Tuple[str, str]]:
        # 返回 时区 的相关信息
        return [(GRPC_HEADER_NAME_TIMEZONE, timezone.get_current_timezone_name())]


trace_unary_unary_client_interceptor = TraceUnaryUnaryClientInterceptor()
language_unary_unary_client_interceptor = LanguageUnaryUnaryClientInterceptor()
timezone_unary_unary_client_interceptor = TimezoneUnaryUnaryClientInterceptor()


def intercept_channel_all(channel):
    """应用所有的拦截器"""
    return grpc.intercept_channel(
        channel,
        trace_unary_unary_client_interceptor,
        language_unary_unary_client_interceptor,
        timezone_unary_unary_client_interceptor,
    )
