# +-------------------------------------------------------------------------
# | Copyright (C) 2024 Toyou, Inc.
# +-------------------------------------------------------------------------
# | Licensed under the Apache License, Version 2.0 (the "License");
# | you may not use this work except in compliance with the License.
# | You may obtain a copy of the License in the LICENSE file, or at:
# |
# | http://www.apache.org/licenses/LICENSE-2.0
# |
# | Unless required by applicable law or agreed to in writing, software
# | distributed under the License is distributed on an "AS IS" BASIS,
# | WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# | See the License for the specific language governing permissions and
# | limitations under the License.
# +-------------------------------------------------------------------------

FROM golang:1.21 as builder
WORKDIR /toyou-csi
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod=vendor  -ldflags "-s -w" -o  _output/toyou-csi-driver ./cmd/disk

FROM debian:stretch
LABEL maintainers="Toyou"
LABEL description="Toyou tyds CSI plugin"

RUN apt-get update && \
    apt-get install -y util-linux e2fsprogs xfsprogs mount ca-certificates udev && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /toyou-csi/_output/toyou-csi-driver /toyou-csi-driver
RUN chmod +x /toyou-csi-driver && \
    mkdir -p /var/log/toyou-csi-driver
ENTRYPOINT ["/toyou-csi-driver"]
