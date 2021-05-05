import pytz
from django.conf import settings
from django.utils import timezone


class LocaleTimezoneMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        # 1. 获取 header 上的时区 ，不存在则使用中国时区
        try:
            current_timezone = pytz.timezone(request.headers['X-Timezone'])
        except KeyError:
            current_timezone = pytz.timezone(settings.TIME_ZONE)

        # 2. 在当前线程激活对应的时区
        timezone.activate(current_timezone)

        response = self.get_response(request)

        # 3. 清除时区
        timezone.deactivate()
        return response
