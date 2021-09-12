import os
from functools import wraps

from celery import Celery
from django.utils import timezone
from django.utils.translation import trans_real

# Set the default Django settings module for the 'task_i18n' program.
from django_i18n import settings

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'django_i18n.settings')

# backend 和 broker 都可以通过环境变量配置，此处为方便直接硬编码
app = Celery('django_i18n', backend='redis://redis/1', broker='redis://redis/0')

# Load task modules from all registered Django apps.
app.autodiscover_tasks()


@app.task(bind=True)
def debug_task(self):
    print(f'Request: {self.request!r}')


def task_i18n(*args, **kwargs):
    """
    支持 i18n 的 Celery 装饰器，
    会在执行 task_func 前将 kwargs 中的语言信息和时区信息激活，
    替换原有的 apply 和 apply_async ，
    在执行它们之前会将语言信息和时区信息放入到 kwargs 中

    :param args: @app.task 的 args 参数
    :param kwargs: @app.task 的 kwargs 参数
    :return: 支持 i18n 的 Celery 装饰器
    """
    def inner(task_func):
        task_func_i18n = _get_task_func_i18n(task_func)
        task = app.task(*args, **kwargs)(task_func_i18n)
        # 替换原有的 apply 和 apply_async
        # 在发送任务前会传入当前线程的语言信息和时区信息到 kwargs 中
        task.apply = _get_apply_i18n(task.apply)
        task.apply_async = _get_apply_i18n(task.apply_async)
        return task

    return inner


KWARGS_LANGUAGE = 'x-language'
KWARGS_TIMEZONE = 'x-timezone'


def _get_task_func_i18n(task_func):
    """
    生成一个支持 i18n 的 task_func ，
    在执行 task_func 前，会将任务的 kwargs 中的语言信息和时区信息激活
    :param task_func:
    :return: 支持 i18n 的 task_func
    """

    @wraps(task_func)
    def task_func_i8n(*args, **kwargs):
        # 获取当前线程原有的语言和时区
        origin_language = trans_real.get_language()
        origin_timezone = timezone.get_current_timezone_name()

        # 从参数中获取传入的 KWARGS_LANGUAGE ，并在当前线程激活对应语言
        language = kwargs.pop(KWARGS_LANGUAGE, settings.LANGUAGE_CODE)
        trans_real.activate(language=language)

        # 从参数中获取传入的 KWARGS_TIMEZONE ，并在当前线程激活对应时区
        timezone.activate(kwargs.pop(KWARGS_TIMEZONE, settings.TIME_ZONE))

        # 执行实际任务
        result = task_func(*args, **kwargs)

        # 在当前线程激活原有的语言和时区
        trans_real.activate(language=origin_language)
        timezone.activate(origin_timezone)

        return result

    return task_func_i8n


def _get_apply_i18n(apply_func):
    """
    生成一个支持 i18n 的 apply_func ，
    在执行 apply_func 前，会将当前线程中的语言信息和时区信息放入到任务的 kwargs 中
    :param apply_func: 
    :return: 支持 i18n 的 apply_func
    """

    def apply_i18n(kwargs=None, **options):
        # 获取 kwargs
        kwargs = kwargs or {}
        # 如果没有 KWARGS_LANGUAGE ，则获取当前线程的传入
        if KWARGS_LANGUAGE not in kwargs:
            kwargs[KWARGS_LANGUAGE] = trans_real.get_language()
        # 如果没有 KWARGS_TIMEZONE ，则获取当前线程的传入
        if KWARGS_TIMEZONE not in kwargs:
            kwargs[KWARGS_TIMEZONE] = timezone.get_current_timezone_name()

        # 执行原有的 apply_func
        return apply_func(kwargs=kwargs, **options)

    return apply_i18n
