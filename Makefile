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

.PHONY: all disk

DISK_IMAGE_NAME=hero74/toyou-csi
DISK_VERSION=v1.0.1
ROOT_PATH=$(pwd)
PACKAGE_LIST=./cmd/... ./pkg/...
CONFIG_FILE=./deploy/kubernetes/base/config.yaml

disk: mod
	docker build -t ${DISK_IMAGE_NAME}-builder:${DISK_VERSION} -f deploy/docker/Dockerfile . --target builder

disk-container:
	docker build -t ${DISK_IMAGE_NAME}:${DISK_VERSION} -f deploy/docker/Dockerfile  .
	docker push ${DISK_IMAGE_NAME}:${DISK_VERSION}

debug:
	docker build -t hero74/toyou-csi:v1.0.1 -f deploy/docker/Dockerfile --cache-from hero74/toyou-csi:v1.0.1 .
	docker push ${DISK_IMAGE_NAME}:${DISK_VERSION}

install:
	kustomize build deploy/kubernetes/overlays/patch|kubectl apply -f -

uninstall:
	kustomize build deploy/kubernetes/overlays/patch|kubectl delete -f -

mod:
	go build ./...

fmt:
	go fmt ${PACKAGE_LIST}

fmt-deep: fmt
	gofmt -s -w -l ./pkg/cloud/ ./pkg/common/ ./pkg/driver ./pkg/rpcserver

sanity-test:
	nohup ${ROOT_PATH}/csi-sanity --csi.endpoint /var/lib/kubelet/plugins/csi.toyou.com/csi.sock -csi.testvolumeexpandsize 21474836480  -ginkgo.noColor &

clean:
	go clean -r -x ./...
	rm -rf ./_output

push:
	docker buildx build -t ${DISK_IMAGE_NAME}:${DISK_VERSION}   --platform=linux/amd64,linux/arm64 -f deploy/docker/Dockerfile . --push