#!/bin/bash

set -e

source $KUDA_CMD_DIR/.config.sh

app_name=$(echo $1 | cut -f1 -d':')
app_version=$(echo $1 | cut -f2 -d':')
app_image="gcr.io/$KUDA_GCP_PROJECT_ID/$app_name"
namespace="default"

# Get cluster's credentials to use kubectl.
gcloud container clusters get-credentials $KUDA_GCP_CLUSTER_NAME

echo "
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: $app_name
  namespace: $namespace
spec:
  template:
    spec:
      containers:
        - image: $app_image
          resources:
            limits:
              nvidia.com/gpu: 1
          volumeMounts:
            - name: secret
              readOnly: true
              mountPath: "/secret"
          env:
            - name: GOOGLE_APPLICATION_CREDENTIALS
              value: /secret/$(basename $KUDA_GCP_CREDENTIALS)
      volumes:
        - name: secret
          secret:
            secretName: $(basename $KUDA_GCP_CREDENTIALS)
" > .kuda-app.k8.yaml

# Cloud Build has a generous free tier is easy enough to use with Skaffold
# So we use it rather than Kaniko.
export GOOGLE_APPLICATION_CREDENTIALS=$KUDA_GCP_CREDENTIALS
# Send version as env variable for Skaffold to use.
export APP_VERSION=$app_version
cat <<EOF | skaffold run -n $namespace -f -
apiVersion: skaffold/v1beta16
kind: Config
build:
  googleCloudBuild:
    projectId: $KUDA_GCP_PROJECT_ID
  artifacts:
    - image: $app_image
  tagPolicy:
    envTemplate:
      template: "{{.IMAGE_NAME}}:{{.APP_VERSION}}"
deploy:
  kubectl:
    manifests:
      - .kuda-app.k8.yaml
EOF

rm .kuda-app.k8.yaml