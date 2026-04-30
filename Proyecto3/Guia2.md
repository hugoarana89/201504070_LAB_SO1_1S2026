# Guía de despliegue — Proyecto 3 (carnet 201504070)

Todo se ejecuta desde **PowerShell** en Windows.

---

## Paso 1 — Habilitar servicios de GCP

```powershell
gcloud services enable container.googleapis.com compute.googleapis.com logging.googleapis.com
```

---

## Paso 2 — Crear el clúster GKE

```powershell
gcloud container clusters create mumnk8s-cluster `
    --zone us-central1-a `
    --num-nodes 2 `
    --machine-type e2-standard-4 `
    --gateway-api standard `
    --disk-size 50 `
    --disk-type pd-standard
```

> ⚠️ Tarda entre 3 y 5 minutos.

---

## Paso 3 — Conectar kubectl al clúster

```powershell
gcloud components install gke-gcloud-auth-plugin

gcloud container clusters get-credentials mumnk8s-cluster --zone us-central1-a
```

Verificar que los nodos están listos:
```powershell
kubectl get nodes
```

Deberías ver 2 nodos en estado `Ready`.

---

## Paso 4 — Instalar certificado de Zot en los nodos de GKE

Los nodos de GKE no conocen tu certificado TLS de Zot.
Este bloque lo instala en todos los nodos automáticamente.

> En la ruta de windows no se reconoce ":" por lo que da problemas al colocar:
> C:\Users\hugo1\.docker\certs.d\34.66.217.216*5000\ca.crt
> C:\Users\hugo1\.docker\certs.d\34.66.217.216:5000\ca.crt
> Lo más rapido es crear una copia de ca.crt en C:\Users\hugo1\.docker\certs.d\34.66.217.216\ca.crt

```powershell
$ZONE = "us-central1-a"
$CERT_PATH = "C:\Users\hugo1\.docker\certs.d\ca.crt"
$NODES = (kubectl get nodes -o jsonpath='{.items[*].metadata.name}') -split ' '

foreach ($NODE in $NODES) {
    Write-Host ">>> Copiando certificado a: $NODE" -ForegroundColor Yellow

    gcloud compute ssh $NODE --zone $ZONE `
        --command "sudo mkdir -p /etc/docker/certs.d/34.66.217.216:5000"

    gcloud compute scp $CERT_PATH "${NODE}:/tmp/ca.crt" --zone $ZONE

    gcloud compute ssh $NODE --zone $ZONE `
        --command "sudo mv /tmp/ca.crt /etc/docker/certs.d/34.66.217.216:5000/ca.crt"

    Write-Host ">>> $NODE OK" -ForegroundColor Green
}
```






```powershell
$ZONE = "us-central1-a"
$NODES = (kubectl get nodes -o jsonpath='{.items[*].metadata.name}') -split ' '

foreach ($NODE in $NODES) {
    Write-Host ">>> Configurando containerd en: $NODE" -ForegroundColor Yellow

    # Copiar el certificado
    gcloud compute scp "C:\Users\hugo1\.docker\certs.d\ca.crt" "${NODE}:/tmp/ca.crt" --zone $ZONE

    # Crear directorio y copiar cert
    gcloud compute ssh $NODE --zone $ZONE --command "sudo mkdir -p /etc/containerd/certs.d/34.66.217.216:5000 && sudo cp /tmp/ca.crt /etc/containerd/certs.d/34.66.217.216:5000/ca.crt"

    # Crear hosts.toml
    gcloud compute ssh $NODE --zone $ZONE --command "echo 'server = ""https://34.66.217.216:5000""' | sudo tee /etc/containerd/certs.d/34.66.217.216:5000/hosts.toml"
    gcloud compute ssh $NODE --zone $ZONE --command "echo '[host.""https://34.66.217.216:5000""]' | sudo tee -a /etc/containerd/certs.d/34.66.217.216:5000/hosts.toml"
    gcloud compute ssh $NODE --zone $ZONE --command "echo '  capabilities = [""pull"", ""resolve""]' | sudo tee -a /etc/containerd/certs.d/34.66.217.216:5000/hosts.toml"
    gcloud compute ssh $NODE --zone $ZONE --command "echo '  ca = ""/etc/containerd/certs.d/34.66.217.216:5000/ca.crt""' | sudo tee -a /etc/containerd/certs.d/34.66.217.216:5000/hosts.toml"

    # Reiniciar containerd
    gcloud compute ssh $NODE --zone $ZONE --command "sudo systemctl restart containerd && echo CONTAINERD_OK"

    Write-Host ">>> $NODE listo" -ForegroundColor Green
}
```


```powershell
gcloud compute ssh gke-mumnk8s-cluster-default-pool-5a8a9046-0tbk --zone us-central1-a --command "cat /etc/containerd/certs.d/34.66.217.216:5000/hosts.toml"
```




Para verificar que el certificado quedó instalado, ejecuta esto en PowerShell — se conecta a cada nodo y muestra el contenido del cert:

```powershell
$ZONE = "us-central1-a"
$NODES = (kubectl get nodes -o jsonpath='{.items[*].metadata.name}') -split ' '

foreach ($NODE in $NODES) {
    Write-Host "`n>>> Verificando nodo: $NODE" -ForegroundColor Yellow
    gcloud compute ssh $NODE --zone $ZONE `
        --command "ls -la /etc/docker/certs.d/34.66.217.216:5000/ && echo '---' && openssl x509 -in /etc/docker/certs.d/34.66.217.216:5000/ca.crt -noout -subject -dates"
}
```
> Se pude borrar la copia en "C:\Users\hugo1\.docker\certs.d\ca.crt"

---

## Paso 5 — Aplicar el namespace

```powershell
cd C:\Users\hugo1\Desktop\201504070_LAB_SO1_1S2026\Proyecto3\k8s

kubectl apply -f 01-namespace.yaml
```

---

## Paso 6 — Crear el secret para Zot

```powershell
$json = '{"auths":{"34.66.217.216:5000":{"auth":""}}}'
$json | Out-File -FilePath "$env:TEMP\dockerconfig.json" -Encoding ascii

kubectl create secret generic zot-registry-secret `
  --from-file=.dockerconfigjson="$env:TEMP\dockerconfig.json" `
  --type=kubernetes.io/dockerconfigjson `
  --namespace=mumnk8s

Remove-Item "$env:TEMP\dockerconfig.json"
```

Verificar que se creó:
```powershell
kubectl get secret zot-registry-secret -n mumnk8s
```

---

El secret está creado correctamente. 

Para verificar conectividad desde los nodos hacia Zot puedes ejecutar esto — lanza un pod temporal que hace el curl desde dentro del clúster:

```powershell
kubectl run zot-test --image=curlimages/curl --restart=Never -n mumnk8s `
  --command -- curl -k https://34.66.217.216:5000/v2/_catalog
```

Espera unos segundos y luego ve los logs:
```powershell
kubectl logs zot-test -n mumnk8s
```

Deberías ver:
```json
{"repositories":["rust-api","go-service","grpc-server","consumer","rabbitmq","valkey"]}
```

Y limpia el pod temporal después:
```powershell
kubectl delete pod zot-test -n mumnk8s
```

---

## Paso 7 — Aplicar los manifiestos de Kubernetes

```powershell
kubectl apply -f 02-deployment-rust.yaml
kubectl apply -f 03-service-rust.yaml
kubectl apply -f 04-gateway.yaml
kubectl apply -f 05-httproute.yaml
kubectl apply -f 06-hpa.yaml
```

---

## Paso 8 — Verificar que el pod levantó

```powershell
kubectl get pods -n mumnk8s -w
```
> Deberías ver el pod pasar de ContainerCreating a Running. Si en 1-2 minutos está en Running sin pasar por ImagePullBackOff, el despliegue base de Rust está funcionando.

> Si el pod queda en `ImagePullBackOff` o `ErrImagePull`, el certificado no quedó bien instalado en los nodos. Ver detalles con:

```powershell
kubectl describe pod -n mumnk8s
```

---

## Paso 9 — Obtener la IP pública del Gateway

```powershell
kubectl get gateway api-gateway -n mumnk8s -w
```

Esperar 3-5 minutos hasta que aparezca una IP en la columna `ADDRESS`.

---

## Paso 10 — Probar con port-forward (antes de usar la IP pública)

En una terminal PowerShell:
```powershell
kubectl port-forward svc/api-rust-service 8080:80 -n mumnk8s
```

En otra terminal PowerShell:
```powershell
curl.exe -X POST http://localhost:8080/grpc-201504070 `
  -H "Content-Type: application/json" `
  -d '{"country":"USA","warplanes_in_air":10,"warships_in_water":5,"timestamp":"2026-03-12T20:15:30Z"}'
```

---

## Paso 11 — Probar con la IP pública del Gateway

```powershell
# Reemplaza con la IP que apareció en el Paso 9
curl.exe -X POST http://<IP-GATEWAY>/grpc-201504070 `
  -H "Content-Type: application/json" `
  -d '{"country":"USA","warplanes_in_air":10,"warships_in_water":5,"timestamp":"2026-03-12T20:15:30Z"}'
```

---

## Paso 12 — Monitorear el HPA en vivo

```powershell
# Terminal 1 — ver escalamiento del HPA
kubectl get hpa api-rust-hpa -n mumnk8s -w

# Terminal 2 — ver pods creándose
kubectl get pods -n mumnk8s -w
```