# Manual Técnico: Gestión de Contenedores en Entornos Virtualizados (KVM)

## 1. Introducción

El presente documento detalla la arquitectura, configuración y despliegue de un sistema distribuido basado en máquinas virtuales Linux (KVM). El proyecto implementa una solución de microservicios contenerizados utilizando **Docker** para la gestión de imágenes y **Containerd** como runtime de ejecución, orquestado mediante un registro de imágenes privado (**Zot**).

## 2. Diagrama de Arquitectura

La solución se basa en un entorno de red privada donde tres máquinas virtuales interactúan para el ciclo de vida de las aplicaciones.

<div align="center">
<img src="img/diagrama.jpg" alt="Diagrama de arquitectura" width="500">
<p><i>Figura 1: Arquitectura del sistema, flujo de imágenes y comunicación entre nodos.</i></p>
</div>

## 3. Descripción de Componentes

### 3.1. Hipervisor y Virtualización

* **KVM (Kernel-based Virtual Machine):** Utilizado como el hipervisor tipo 2 para la creación de máquinas virtuales.
* **QEMU/Libvirt:** Herramientas de emulación y gestión de la plataforma de virtualización.
* **Red Virtual (Bridge):** Se configuró un puente virtual para permitir la comunicación interna entre las VMs bajo el segmento `192.168.122.0/24`.

### 3.2. Máquinas Virtuales (Nodos)

| VM | Rol | Componentes Clave | IP Estática |
| --- | --- | --- | --- |
| **VM1** | Nodo de Ejecución A | Containerd, API 1, API 2 | `192.168.122.101` |
| **VM2** | Nodo de Ejecución B | Containerd, API 3 | `192.168.122.102` |
| **VM3** | Nodo de Gestión | Docker, Zot Registry | `192.168.122.103` |

### 3.3. Tecnologías de Contenedores

* **Docker:** Empleado exclusivamente en la VM de gestión para la construcción (*build*) y carga (*push*) de imágenes.
* **Zot Registry:** Registro de contenedores OCI-native que almacena las imágenes de las APIs para ser consumidas por los nodos.
* **Containerd:** Runtime de alto rendimiento utilizado en los nodos de ejecución para el despliegue de contenedores mediante la herramienta de línea de comandos `ctr`.

## 4. Configuración del Sistema

### 4.1. Configuración de Red (Netplan)

Para garantizar la persistencia de las conexiones, se inhabilitó DHCP en las VMs y se definieron archivos de configuración YAML en `/etc/netplan/`, asegurando que el Gateway apunte a la interfaz del host virtual (`192.168.122.1`).

### 4.2. Especificaciones de las APIs

Las APIs fueron desarrolladas en el lenguaje **Go**, cada una con su respectivo `Dockerfile` basado en imágenes ligeras (Alpine) para optimizar el almacenamiento y el tiempo de transferencia en la red.

## 5. Flujo de Funcionamiento Técnico

1. **Construcción:** El código fuente se transfiere a la VM3. Docker construye las imágenes etiquetándolas con el formato `localhost:5000/api-nombre:tag`.
2. **Distribución:** Las imágenes se cargan en el registro Zot en el puerto `5000`.
3. **Despliegue:** * VM1 y VM2 consultan el catálogo de VM3 vía HTTP.
* `Containerd` realiza un *pull* de las imágenes necesarias.


4. **Ejecución:** Los contenedores se ejecutan con la bandera `--net-host`, permitiendo que las APIs utilicen la pila de red de la VM para comunicarse entre ellas mediante las IPs estáticas configuradas.

## 6. Resolución de Problemas (Troubleshooting)

* **Error de conexión al Registro:** Verificar que el servicio Zot esté corriendo en VM3 (`docker ps`) y que el firewall permita tráfico en el puerto 5000.
* **Fallo de Pull en Containerd:** Asegurarse de usar la bandera `--plain-http` si el registro no cuenta con certificados SSL/TLS configurados.
* **Comunicación entre APIs:** Validar el estado de las APIs mediante el endpoint `/health` y verificar que las rutas en el código apunten a las IPs correctas de las otras VMs.

---

**Desarrollado para:** Facultad de Ingeniería, USAC.
**Curso:** Sistemas Operativos 1.