## 简介

我们前面通过四篇文章介绍了如何在 Django 和 Gin 中实现全链路追踪和国际化与本地化。

- [Django i18n](demo%20(simple%20i18n).md): Django 中的国际化与本地化
- [Gin i18n](../gin/demo%20(simple%20i18n).md): Gin 中的国际化与本地化
- [gRPC APM](../demo%20(apm).md): gRPC 的全链路追踪
- [gRPC i18n](../demo%20(gRPC%20i18n).md): gRPC 的全链路国际化与本地化

本次我们将介绍利用 `monkey patch` 方法，实现 Celery 上无侵入的全链路国际化与本地化。

## 背景

在我们的实际使用场景中，不仅需要实现微服务上的全链路国际化与本地化，还需要在 Celery 中也实现类似的全链路逻辑。

例如：客户需要导出一些数据，导出任务不能很快处理完，所以需要通过 Celery 处理。而导出文件的文件名和固定信息等都需要根据客户浏览器的语言自动适配，并且时间字段也需要根据客户所在时区自动转换。

我们已经实现了无侵入式的微服务上全链路传递信息，自然也想在 Celery 中做到无侵入传递所需信息。

但 Celery 并不支持中间件或者拦截器的逻辑，所以我们无法直接借助框架完成我们的目的。

不过 Python 是动态语言，我们可以借助 `monkey patch` 的方法修改 Celery 注册任务的逻辑，替换任务的发送和执行逻辑。

这可以让我们在发送任务前将所需信息放入 `kwargs` 中当作参数，在执行任务前从 `kwargs` 中获取传递的信息，不会影响实际执行的任务函数。

## 确定修改逻辑

我们先看看 Celery 是如何注册并执行任务的：

```python
# 创建 celery 实例
# backend 和 broker 都可以通过环境变量配置，此处为方便直接硬编码
app = Celery('django_i18n', backend='redis://redis/1', broker='redis://redis/0')

# 从已注册的应用中自动发现任务
app.autodiscover_tasks()

# 在任务上加上 @app.task() 就可以使其变成任务并被自动发现
@app.task()
def hello(username: str) -> str:
    return f'Hello, {username}.'

# 同步执行，会在当前线程中执行 hello 任务
hello.apply(args=('idealism',))
# 异步执行，会通过 app 中指定的 backend 和 broker 交给 worker 执行
hello.apply_async(kwargs=dict(username='idealism'))
```

我们可以发现注册任务是通过装饰器 `@app.task()` 完成的，而执行任务是通过 `apply` 和 `apply_async` 完成的。

那么我们的目标就很明确：实现一个 `@app.task()` 的 i18n 版本，修改任务的发送和执行两块逻辑。

1. 替换任务的 `apply` 和 `apply_async` 两个函数，让它们在执行原有发送逻辑前，将当前线程的语言信息和时区信息放入到 `kwargs` 中当作参数传递给 worker 
2. 替换任务函数本身，在执行原本的任务函数前，从 `kwargs` 中去除语言信息和时区信息，并在当前线程中激活

## 修改发送任务逻辑

我们定义 `_get_apply_i18n` 函数，它接受一个函数 `apply_func` ，并返回一个支持 i18n 的发送函数。

该函数在被调用时会先从当前线程中获取语言信息和时区信息（没有则使用默认的），然后放入到 `kwargs` 中。

有一点值得注意的是：这里只有在 `kwargs` 中没有的相应信息才会放入。

这样支持让调用者指定一个固定的语言和时区，在某些特殊的场景下也可以发挥作用。

```python
KWARGS_LANGUAGE = 'x-language'
KWARGS_TIMEZONE = 'x-timezone'


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
```

## 修改执行任务逻辑

我们定义 `_get_task_func_i18n` 函数，它接受一个函数 `task_func` ，并返回一个支持 i18n 的任务函数。

该函数在被调用时会先保存当前线程的语言和时区，然后从 `kwargs` 中获取语言信息和时区信息（没有则使用默认的），并在当前线程中激活，再执行原有的任务函数，最后恢复原有的语言和时区。

```python
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
```

## 修改注册任务逻辑

我们现在可以获取到支持 i18n 的发送函数和任务函数，接下来就是在注册任务时替换原有的函数即可。

```python
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
```

## 使用支持 i18n 的任务注册装饰器

现在我们已经有了支持 i18n 的任务注册装饰器，那么直接替换原有的 `@app.task()` 即可，做到了无侵入式传递语言信息和时区信息。

```python
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
```

同时，我写了一个 `docker-compose` 的文件 `i18n.yml` ，可以很方便地启动所需的所有服务： redis, django-i18n, celery-i18n, gin-i18n ，可以直接在对应的仓库中查看。

在根目录运行 `docker-compose -f i18n.yml up` 即可启动全部应用，并能直接在控制台查看所有应用的日志。

## 测试

访问 `http://127.0.0.1:8000/hello-with-celery-apply/idealism/` 可以测试同步任务。

访问 `http://127.0.0.1:8000/hello-with-celery-apply-async/idealism/` 可以测试异步任务。

同步任务会直接在当前线程执行，所以在 celery-i18n 应用中没有任务执行的记录。

而异步任务会通过 redis 在 worker 执行，能看到 celery-i18n 应用中的日志如下：

```shell
celery-i18n    | [2021-09-12 11:23:17,721: INFO/MainProcess] Task django_i18n.task.hello[f67ea302-4585-4f32-b51b-df46aeb8de20] received
celery-i18n    | [2021-09-12 11:23:17,741: INFO/ForkPoolWorker-7] Task django_i18n.task.hello[f67ea302-4585-4f32-b51b-df46aeb8de20] succeeded in 0.01851999998325482s: 'Current Language: en-us<br/>Hello, idealism.<br/>(Asia/Shanghai) 2021-09-12 19:23:17.739392+08:00'
```

不同请求头测试结果如下（可以使用 [ModHeader](https://chrome.google.com/webstore/detail/modheader/idgpnmonknjnojddfkpgkljpfnnfcklj) 插件在浏览器中快速修改请求头）：

```
# Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
# x-timezone: 
当前语言: zh-hans
你好，idealism。
(UTC) 2021-09-12 11:24:28.778792+00:00

# Accept-Language: en
# x-timezone: Europe/Berlin
Current Language: en-us
Hello, idealism.
(Europe/Berlin) 2021-09-12 13:24:47.783920+02:00

# Accept-Language: en
# x-timezone: Asia/Shanghai
Current Language: en-us
Hello, idealism.
(Asia/Shanghai) 2021-09-12 19:25:05.392891+08:00
```

这个简单的测试说明 Django 中能从请求头中正确获取到语言和时区信息，并能正确通过我们刚刚实现的支持 i18n 的 Celery 装饰器传递给 worker 。

## 小结

本文主要讲了如何实现 Celery 上无侵入的全链路国际化与本地化。

主要运用 `monkey patch` 方法，使得我们在运行时可以动态修改函数的逻辑，让我们在注册、发送、执行任务前运行所需的逻辑。

这种无侵入式的方法提供了一种简单的改造方式，方便后续传递其他信息，极大地降低了接入改造的成本。

从最开始应用内的 i18n 到 gRPC 上的全链路追踪，再到 gRPC 上的全链路国际化与本地化，最后到 Celery 上的全链路国际化与本地化。

可以发现这些应用场景听起来很高大上，但使用的技术都很简单，关键在于充分利用框架和语言本身的特性，并学会从类似的功能中举一反三，迁移到我们需要的领域中。
