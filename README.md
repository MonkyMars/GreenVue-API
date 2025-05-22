# GreenVue API

Backend API for GreenVue.eu - an EU-focused sustainable marketplace for selling pre-owned items.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2.52.6-brightgreen.svg)](https://gofiber.io/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 🌱 Overview

GreenVue is a sustainable marketplace platform focused on the EU market, designed to facilitate the buying and selling of pre-owned items. This repository contains the backend API that powers both the [mobile app](https://github.com/MonkyMars/GreenVue-Mobile) and [web platform](https://github.com/MonkyMars/GreenVue-Web).

Our mission is to promote sustainability by extending the lifecycle of products and reducing waste through a user-friendly marketplace that encourages reuse over disposal.

## 🛠️ Tech Stack

- **Language**: Go 1.24
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [Supabase](https://supabase.io/)
- **Image Storage**: Supabase Storage Buckets
- **Authentication**: JWT (JSON Web Tokens)
- **API Documentation**: (TBD)
- **Containerization**: Docker
- **Hosting**: Railway

## ✨ Features

- **Authentication**

  - Registration and login
  - JWT-based authentication
  - Refresh token support

- **Listings Management**

  - Create, retrieve and delete listings
  - Category-based listing organization
  - Image upload support (WebP conversion, resizing, async processing)
  - Seller-specific listings

- **Background Jobs**

  - Scheduled tasks using the same API port
  - Job management through REST API
  - Configurable job intervals
  - Built-in jobs for common tasks (cleanup, notifications, email sending)
  - Asynchronous image processing for better performance
  - [Job System Documentation](docs/jobs.md)
  - [Image Processing Documentation](docs/images.md)

- **User Profiles**

  - User information management
  - Seller profiles

- **Chat System**

  - Real-time messaging via WebSockets
  - Conversation management
  - Message history

- **Reviews & Ratings**

  - Post and retrieve product reviews

- **Favorites**

  - Save and manage favorite listings

- **Security**
  - Rate limiting
  - CORS protection
  - Input validation
  - Structured error handling

## 🔒 Security

GreenVue takes security seriously. Our security practices include:

- Environment variable management for credentials
- Secure JWT implementation
- Rate limiting to prevent abuse
- Input validation and sanitization
- Regular dependency updates

For vulnerability reporting, please see our [security policy](SECURITY.md).

## 🗺️ Roadmap

Future development plans:

- User profile management improvements
- Password reset functionality
- Improving Email verification system
- Search API with filtering capabilities
- Pagination for listing results
- Order management system
- Purchase history endpoints
- Enhanced review system

## 🔗 Related Projects

- [GreenVue Mobile App](https://github.com/MonkyMars/GreenVue-Mobile)
- [GreenVue Web Platform](https://github.com/MonkyMars/GreenVue-Web)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
