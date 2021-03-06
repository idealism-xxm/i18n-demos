from django.http import HttpResponse
from django.utils import timezone
from django.utils.translation import get_language
from django.utils.translation import ugettext as _t
from django_i18n.constants import CURRENT_LANGUAGE
from django_i18n.constants import HELLO
from django_i18n import task

import grpccli.service


# Create your views here.
def hello(request, username):
    # 获取本次请求的时区
    cur_timezone = timezone.get_current_timezone()
    return HttpResponse(
        f"{_t(CURRENT_LANGUAGE).format(cur_lang=get_language())}<br/>"
        f"{_t(HELLO).format(username=username)}<br/>"
        # timezone.now() 返回的时间时区是 settings.TIME_ZONE
        f"({cur_timezone}) {timezone.now().astimezone(cur_timezone)}"
    )


def hello_with_grpc(request, username):
    return HttpResponse(grpccli.service.hello(username))


def hello_with_celery_apply(request, username):
    result = task.hello.apply(args=(username,))
    return HttpResponse(result.get())


def hello_with_celery_apply_async(request, username):
    result = task.hello.apply_async(kwargs=dict(username=username))
    return HttpResponse(result.get())
