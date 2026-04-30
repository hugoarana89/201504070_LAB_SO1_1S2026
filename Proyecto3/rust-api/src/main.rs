use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use serde::{Deserialize, Serialize};
use log::{info, error};

// Estructura del mensaje que llega desde Locust
#[derive(Deserialize, Serialize, Debug)]
struct WarReport {
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: String,
}

// Estructura de respuesta
#[derive(Serialize)]
struct ApiResponse {
    status: String,
    message: String,
}

// Handler del endpoint principal
async fn receive_report(
    report: web::Json<WarReport>,
    go_client: web::Data<reqwest::Client>,
    go_url: web::Data<String>,
) -> HttpResponse {
    info!("Reporte recibido: {:?}", report);

    // Validaciones básicas
    if report.country.len() != 3 {
        return HttpResponse::BadRequest().json(ApiResponse {
            status: "error".to_string(),
            message: "El país debe tener exactamente 3 letras".to_string(),
        });
    }

    if report.warplanes_in_air < 0 || report.warplanes_in_air > 50 {
        return HttpResponse::BadRequest().json(ApiResponse {
            status: "error".to_string(),
            message: "warplanes_in_air debe estar entre 0 y 50".to_string(),
        });
    }

    if report.warships_in_water < 0 || report.warships_in_water > 30 {
        return HttpResponse::BadRequest().json(ApiResponse {
            status: "error".to_string(),
            message: "warships_in_water debe estar entre 0 y 30".to_string(),
        });
    }

    // Reenviar al Deployment de Go
    let go_endpoint = format!("{}/report", go_url.get_ref());
    
    match go_client
        .post(&go_endpoint)
        .json(&report.into_inner())
        .send()
        .await
    {
        Ok(response) => {
            if response.status().is_success() {
                info!("Reporte enviado exitosamente a Go");
                HttpResponse::Ok().json(ApiResponse {
                    status: "ok".to_string(),
                    message: "Reporte recibido y enviado a procesamiento".to_string(),
                })
            } else {
                error!("Go respondió con error: {}", response.status());
                HttpResponse::InternalServerError().json(ApiResponse {
                    status: "error".to_string(),
                    message: "Error al procesar el reporte".to_string(),
                })
            }
        }
        Err(e) => {
            error!("Error conectando a Go: {}", e);
            HttpResponse::InternalServerError().json(ApiResponse {
                status: "error".to_string(),
                message: "Error de conexión con el servicio de procesamiento".to_string(),
            })
        }
    }
}

// Health check
async fn health() -> HttpResponse {
    HttpResponse::Ok().json(ApiResponse {
        status: "ok".to_string(),
        message: "API Rust funcionando".to_string(),
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::default().default_filter_or("info"));

    // URL del servicio Go, configurable por variable de entorno
    let go_service_url = std::env::var("GO_SERVICE_URL")
        .unwrap_or_else(|_| "http://localhost:8081".to_string());

    info!("Iniciando API Rust en puerto 8080");
    info!("Go service URL: {}", go_service_url);

    let go_url = web::Data::new(go_service_url);
    let http_client = web::Data::new(reqwest::Client::new());

    HttpServer::new(move || {
        App::new()
            .app_data(go_url.clone())
            .app_data(http_client.clone())
            // Configurar tamaño máximo del body para alta carga
            .app_data(web::JsonConfig::default().limit(1024 * 1024))
            .wrap(middleware::Logger::default())
            .route("/grpc-201504070", web::post().to(receive_report))
            .route("/", web::get().to(health)) // tiene que ser / no /health para que funcione con el gateway api
    })
    .workers(4)                    // workers para manejar alta carga
    .bind("0.0.0.0:8080")?
    .run()
    .await
}