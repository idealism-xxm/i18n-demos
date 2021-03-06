"""django_i18n URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/3.2/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""
from django.contrib import admin
from django.urls import path
from django.urls import re_path

from . import views

urlpatterns = [
    path('admin/', admin.site.urls),
    re_path(r'hello/(?P<username>\w+)/', views.hello),
    re_path(r'hello-with-grpc/(?P<username>\w+)/', views.hello_with_grpc),
    re_path(r'hello-with-celery-apply/(?P<username>\w+)/', views.hello_with_celery_apply),
    re_path(r'hello-with-celery-apply-async/(?P<username>\w+)/', views.hello_with_celery_apply_async),
]
