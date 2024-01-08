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

# Builder stage
FROM golang:1.21 as builder
WORKDIR /toyou-csi
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod=vendor -ldflags "-s -w" -o _output/toyou-csi-driver ./cmd/disk

# Final image stage
FROM centos:7
LABEL maintainers="Toyou"
LABEL description="Toyou tyds CSI plugin"

# Install required packages
RUN yum install -y file which e2fsprogs xfsprogs e2fsprogs-libs multipath-tools sg3_utils device-mapper-multipath lsscsi

# Copy the built binary and configuration files
COPY --from=builder /toyou-csi/_output/toyou-csi-driver /toyou-csi-driver
COPY deploy/kubernetes/base/config.yaml /etc/config/config.yaml

# Create required directories and set permissions


# Set the entrypoint
ENTRYPOINT ["/toyou-csi-driver"]
