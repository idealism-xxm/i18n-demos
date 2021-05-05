FROM python:3.8

WORKDIR /idealism-xxm/django

COPY . .

RUN apt-get update && \
    apt-get install gettext -y

RUN pip install poetry && \
    # disable virtual env in docker
    poetry config virtualenvs.create false && \
    poetry install

# compile messages
RUN python manage.py compilemessages

CMD [ "python3.8", "manage.py", "runserver", "0.0.0.0:8000" ]
