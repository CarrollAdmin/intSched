package constants

const DeploymentTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .deploymentName }}
  namespace: slurm
  labels:
    app: Slurm
  annotations:
    flag: "0"
    cpu-bandwidth: "{{ .totalCPUSPerNode }}"
    app: SLURM-JOB
    type: {{ .jobType }}
    dynamic-resource-management: "1"
spec:
  replicas: {{ .nodeNum }}
  selector:
    matchLabels:
      app: Slurm
  template:
    metadata:
      annotations:
        app: SLURM-JOB
        schedulerName: xsched
      labels:
        app: Slurm
    spec:
      schedulerName: xsched
      containers:
      - name: testdeployment
        image: wardsco/sleep:latest
        imagePullPolicy: IfNotPresent
`
