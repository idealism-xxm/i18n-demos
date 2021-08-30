## 简介

在 [Django i18n](django/demo%20(simple%20i18n).md) 和 [Gin i18n](gin/demo%20(simple%20i18n).md) 中分别介绍了如何在 Django 和 Gin 中实现国际化和本地化，并在 [gRPC APM](demo%20(apm).md) 中介绍了如何实现 gRPC 的全链路追踪。

现在我们就要组合前面介绍的知识，实现我们所需要的全链路国际化和本地化功能（当然可以扩展至任意需要全链路传递的信息）。

本次将实现 gRPC 和 Celery 上的全链路国际化，全链路追踪和 Elasticsearch + Apm + Kibana 环境搭建可以直接查看 [gRPC APM](demo%20(apm).md) 。
