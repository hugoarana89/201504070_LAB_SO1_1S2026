# 📦 Explicación del Proyecto: Arquitectura con Docker, Containerd y Registro Privado

---

## 1️⃣ Diferencias entre Docker y Containerd

Docker y containerd no son equivalentes, aunque están relacionados.

- **Docker** es una herramienta de alto nivel, más simple y orientada al usuario.
- **Containerd** es una herramienta de más bajo nivel, enfocada únicamente en la ejecución de contenedores.

### 🔹 Comandos comunes en Docker

```bash
docker build
docker run
docker push
docker pull
docker images
````

Docker permite construir, gestionar y publicar imágenes fácilmente.

### 🔹 Comandos comunes en Containerd

```bash
ctr images pull
ctr images push
ctr images list
ctr run
ctr containers list
```

Containerd permite descargar y ejecutar imágenes, pero **no está diseñado para construir imágenes de manera sencilla** como Docker.

### ✅ Conclusión operativa

En este proyecto:

* Docker se utiliza para **construir imágenes**
* Zot se utiliza para **almacenarlas**
* Containerd se utiliza para **ejecutarlas**

---

# 🧠 2️⃣ Idea General del Proyecto

El proyecto implementa una arquitectura distribuida basada en tres máquinas virtuales:

| VM  | Rol                                           |
| --- | --------------------------------------------- |
| VM1 | Ejecuta API1 y API2                           |
| VM2 | Ejecuta API3                                  |
| VM3 | Actúa como registro privado de imágenes (Zot) |

---

## 🔥 ¿Qué es Zot?

Zot es un **registro privado de contenedores**.

Su función es almacenar imágenes Docker para que otras máquinas dentro de la red puedan descargarlas.
Cumple el mismo propósito que DockerHub, pero dentro de una red privada.

---

# 🔁 3️⃣ Flujo Completo del Proyecto

A continuación se describe la secuencia completa del funcionamiento del sistema.

---

## 🥇 Paso 1 — Desarrollo de las APIs

Se desarrollaron tres servicios independientes:

* API1
* API2
* API3

Cada uno expone un endpoint `/health` para validación de estado.

---

## 🥈 Paso 2 — Construcción de Imágenes Docker (VM3)

En la máquina virtual VM3 se construyen las imágenes:

```bash
cd api1
docker build -t api1-201504070 .

cd ../api2
docker build -t api2-201504070 .

cd ../api3
docker build -t api3-201504070 .
```

Resultado: tres imágenes listas para distribución.

---

## 🥉 Paso 3 — Publicación en el Registro Privado (Zot en VM3)

Las imágenes se etiquetan para apuntar al registro privado:

```bash
docker tag api1-201504070 192.168.122.103:5000/api1-201504070
docker tag api2-201504070 192.168.122.103:5000/api2-201504070
docker tag api3-201504070 192.168.122.103:5000/api3-201504070
```

Posteriormente se publican:

```bash
docker push 192.168.122.103:5000/api1-201504070
docker push 192.168.122.103:5000/api2-201504070
docker push 192.168.122.103:5000/api3-201504070
```

En este punto, VM3 funciona como repositorio privado de imágenes.

---

## 🏗 Paso 4 — Descarga de Imágenes en VM1 y VM2

En VM1:

```bash
sudo ctr images pull 192.168.122.103:5000/api1-201504070
sudo ctr images pull 192.168.122.103:5000/api2-201504070
```

En VM2:

```bash
sudo ctr images pull 192.168.122.103:5000/api3-201504070
```

Las máquinas no necesitan el código fuente, solo las imágenes.

---

## 🚀 Paso 5 — Ejecución de Contenedores

En VM1:

```bash
sudo ctr run -d -p 8080:8080 192.168.122.103:5000/api1-201504070 api1
sudo ctr run -d -p 8081:8080 192.168.122.103:5000/api2-201504070 api2
```

En VM2:

```bash
sudo ctr run -d -p 8080:8080 192.168.122.103:5000/api3-201504070 api3
```

Ahora los servicios están en ejecución.

---

## 🔄 Paso 6 — Comunicación entre APIs

El flujo de comunicación es el siguiente:

1. El usuario accede a API1
2. API1 realiza una solicitud a API2
3. API2 realiza una solicitud a API3
4. API3 responde
5. Cada API valida el endpoint `/health`

Esto demuestra comunicación entre microservicios mediante REST.

---

# 🧩 4️⃣ Arquitectura Final del Sistema

## 📌 Separación de Responsabilidades

* **Docker** → Construcción de imágenes
* **Zot** → Almacenamiento de imágenes
* **Containerd** → Ejecución de contenedores
* **APIs** → Comunicación entre servicios

---

## 🖥 Distribución de Infraestructura

```
Código fuente
      ↓
Docker (VM3)
      ↓
Imagen creada
      ↓
Zot (VM3 - Registro privado)
      ↓
--------------------------------
↓                              ↓
VM1                           VM2
(containerd)                  (containerd)
(API1, API2)                  (API3)
```

---

# 🎯 5️⃣ Objetivos Técnicos Demostrados

El proyecto demuestra que se comprende:

1. La diferencia entre código fuente e imagen de contenedor
2. El uso de un registro privado
3. La distribución de imágenes entre máquinas virtuales
4. La ejecución de contenedores con containerd
5. La comunicación entre microservicios

---

# 📌 Consideración Importante

Se recomienda construir las imágenes directamente en VM3, donde se encuentra el registro Zot.

Construirlas en otra máquina puede requerir configuraciones adicionales del registro como inseguro.

---

# 🏁 Conclusión

El sistema implementa una arquitectura distribuida realista donde:

* El desarrollo se desacopla de la ejecución.
* Las imágenes se centralizan en un registro privado.
* Las máquinas de producción solo ejecutan contenedores.
* Los servicios se comunican entre sí mediante HTTP.

Este enfoque replica el funcionamiento de entornos empresariales modernos basados en contenedores.

---
