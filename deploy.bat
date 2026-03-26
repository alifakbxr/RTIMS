@echo off
REM RTIMS Production Deployment Script for Windows
REM This script helps deploy RTIMS to production

echo ======================================
echo RTIMS Production Deployment
echo ======================================

REM Check if .env.production exists
if not exist .env.production (
    echo Error: .env.production file not found!
    echo Please copy .env.production.example to .env.production and fill in your values.
    exit /b 1
)

REM Create necessary directories
echo Creating necessary directories...
if not exist nginx\ssl mkdir nginx\ssl
if not exist nginx\logs mkdir nginx\logs

REM Pull latest images
echo Pulling latest Docker images...
docker-compose -f docker-compose.production.yml pull

REM Start services
echo Starting production services...
docker-compose -f docker-compose.production.yml up -d

REM Wait for services to be healthy
echo Waiting for services to start...
timeout /t 10 /nobreak

REM Check service health
echo Checking service health...
docker-compose -f docker-compose.production.yml ps

echo.
echo ======================================
echo Deployment Complete!
echo ======================================
echo.
echo To view logs:
echo   docker-compose -f docker-compose.production.yml logs -f
echo.
echo To stop services:
echo   docker-compose -f docker-compose.production.yml down
echo.
