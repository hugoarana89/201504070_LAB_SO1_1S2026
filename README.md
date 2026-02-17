# Gestión de Microservicios en Entornos Virtualizados (KVM) 🚀

Este proyecto demuestra el despliegue de una arquitectura de microservicios utilizando virtualización nativa con **KVM**, gestión de imágenes mediante un registro privado (**Zot**) y ejecución de contenedores con **Containerd**.

## 📌 Tabla de Contenidos
- [Arquitectura](#arquitectura)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Requisitos](#requisitos)
- [Configuración de Red](#configuración-de-red)
- [Despliegue](#despliegue)
- [Uso de APIs](#uso-de-apis)

## 🏗 Arquitectura
El sistema se compone de tres nodos virtuales interconectados en una red privada:

1.  **VM1 (Nodo de Ejecución):** Ejecuta la API 1 y API 2 mediante `containerd`.
2.  **VM2 (Nodo de Ejecución):** Ejecuta la API 3 mediante `containerd`.
3.  **VM3 (Nodo de Gestión):** Actúa como registro privado utilizando `Zot` y construye las imágenes con `Docker`.



## 📂 Estructura del Proyecto
```text
201504070_LAB_SO1_1S2026/
├── api1/
│   ├── main.go         # Servicio principal (Punto de entrada)
│   ├── Dockerfile      # Configuración de imagen
│   └── go.mod
├── api2/
│   ├── main.go
│   ├── Dockerfile
│   └── go.mod
├── api3/
│   ├── main.go
│   ├── Dockerfile
│   └── go.mod
├── docs/
│   ├── manual_tecnico.pdf
│   └── guia_instalacion.md
└── README.md

```

## 🛠 Requisitos

* Linux con soporte para virtualización KVM.
* `virt-manager` y `libvirt`.
* Docker (en VM3).
* Containerd (en VM1 y VM2).
* Go 1.21+ (para desarrollo local).

## 🌐 Configuración de Red

Se utilizan IPs estáticas para garantizar la comunicación entre servicios:

* **VM1:** `192.168.122.101`
* **VM2:** `192.168.122.102`
* **VM3:** `192.168.122.103` (Registro Zot en puerto 5000)

## 🚀 Despliegue

### 1. Preparación de Imágenes (VM3)

Desde la carpeta raíz del proyecto en la VM3:

```bash
# Construir imágenes
docker build -t api1-201504070:latest ./api1
docker build -t api2-201504070:latest ./api2
docker build -t api3-201504070:latest ./api3

# Etiquetar y subir al registro local
docker tag api1-201504070:latest localhost:5000/api1-201504070:latest
docker push localhost:5000/api1-201504070:latest
# (Repetir para api2 y api3)

```

### 2. Ejecución en Nodos (VM1 y VM2)

En los nodos de ejecución, se utiliza `ctr` para descargar y correr las imágenes:

```bash
# Ejemplo en VM1
sudo ctr image pull --plain-http 192.168.122.103:5000/api1-201504070:latest
sudo ctr run -d --net-host 192.168.122.103:5000/api1-201504070:latest api1

```

## 📡 Uso de APIs

Las APIs se comunican de forma encadenada para validar el estado del sistema.

* **API 1 (Puerto 8080):** Recibe peticiones y consulta a API 2.
* **API 2 (Puerto 8081):** Procesa datos y consulta a API 3.
* **API 3 (Puerto 8082):** Responde con el estado final del sistema.

### Validación de Salud

Cada API expone un endpoint `/health` para monitoreo:

```bash
curl [http://192.168.122.101:8080/health](http://192.168.122.101:8080/health)

```

---

**Curso:** Sistemas Operativos 1 - USAC

```

```
