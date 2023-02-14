ARG base_image
FROM ${base_image}

USER root

ARG build_id=0
RUN echo ${build_id}

RUN microdnf -y module enable nodejs:18
RUN microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install -y make gcc gcc-c++ libatomic_ops git openssl-devel nodejs npm nodejs-nodemon nss_wrapper which && microdnf clean all

RUN echo uid:gid "1000:1000"
USER 1000:1000

RUN echo "CNB_STACK_ID: "