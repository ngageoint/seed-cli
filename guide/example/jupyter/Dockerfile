FROM alpine:3.12

RUN apk update && \
    apk add --no-cache build-base libzmq musl-dev zeromq-dev python3 python3-dev py-pip jpeg-dev zlib-dev

RUN pip install --upgrade pip && \
    pip install nbconvert && \
    pip install papermill && \
    pip install jupyter_client && \
    pip install Pillow >= 7 && \
    python3 -m pip install ipykernel && \
    python3 -m ipykernel install --user

WORKDIR /usr/algo

COPY ./run_notebook.py .
COPY ./image_rotate.sh .
COPY ./image_rotate.ipynb .
RUN chmod +x ./image_rotate.sh