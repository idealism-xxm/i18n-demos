from django.utils import timezone
from django.utils.translation import get_language
from django.utils.translation import ugettext as _t

from django_i18n.celery import task_i18n
from django_i18n.constants import CURRENT_LANGUAGE
from django_i18n.constants import HELLO


@task_i18n()
def hello(username: str) -> str:
    # 获取本次请求的时区
    cur_timezone = timezone.get_current_timezone()
    return (
        f"{_t(CURRENT_LANGUAGE).format(cur_lang=get_language())}<br/>"
        f"{_t(HELLO).format(username=username)}<br/>"
        # timezone.now() 返回的时间时区是 settings.TIME_ZONE
        f"({cur_timezone}) {timezone.now().astimezone(cur_timezone)}"
    )
