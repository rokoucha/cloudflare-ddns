FROM python:3.9-buster as builder

WORKDIR /opt/app

COPY requirements.txt /opt/app
RUN pip3 install -r requirements.txt

FROM python:3.9-slim-buster as runner

WORKDIR /opt/app

COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages

COPY cloudflare-ddns.py /opt/app

ENTRYPOINT ["python", "/opt/app/cloudflare-ddns.py"]
