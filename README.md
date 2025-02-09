# NextGenBuy is a E-Commerce Platform

## 🚀 Overview
This is a scalable and modern e-commerce platform that handles both customers and merchants. It includes features like product management, offers & discounts, shipping options, and secure authentication.

## 📂 Tech Stack
- **Backend:** GoLang (Gin framework)
- **Database:** PostgreSQL
- **ORM Model:** GORM
- **Authentication:** JWT-based authentication

## ⚡ Features
- **Merchant Dashboard:** Add, update, and manage products
- **Product Categories & Offers:** Discounts, flash sales, coupon codes
- **Customer Account Management:** Login, order tracking, cart management
- **Payment Integration:** Secure payments via Stripe/PayPal/Juspay
- **Order Tracking & Shipping:** Real-time order status updates

## 🛠 Setup Instructions
```bash
# Clone the repository
git clone https://github.com/yourusername/ecommerce-project.git
cd ecommerce-project

# Install dependencies
yarn install  # For frontend
go mod tidy   # For backend

# Set up environment variables
cp .env.example .env

# Run the backend
go run main.go

# Run the frontend
yarn start
```

## 📌 API Endpoints
### 🔑 Authentication
| Method | Endpoint       | Description |
|--------|--------------|-------------|
| POST   | `/auth/signup`  | User registration |
| POST   | `/auth/login`  | User login & token generation |
| GET    | `/auth/profile` | Get logged-in user details |

### 🛍 Products
| Method | Endpoint               | Description |
|--------|----------------------|-------------|
| GET    | `/products`            | Get all products |
| GET    | `/products/:id`        | Get product by ID |
| POST   | `/products` (Merchant) | Add a new product |
| PUT    | `/products/:id` (Merchant) | Update product details |
| DELETE | `/products/:id` (Merchant) | Remove a product |

### 🏷 Offers & Discounts
| Method | Endpoint               | Description |
|--------|----------------------|-------------|
| GET    | `/offers`            | Get all active offers |
| POST   | `/offers` (Merchant) | Create a new offer |
| PUT    | `/offers/:id` (Merchant) | Update offer details |
| DELETE | `/offers/:id` (Merchant) | Remove an offer |

### 🛒 Cart & Checkout
| Method | Endpoint               | Description |
|--------|----------------------|-------------|
| GET    | `/cart`              | Get cart details |
| POST   | `/cart/add`          | Add item to cart |
| DELETE | `/cart/remove/:id`   | Remove item from cart |
| POST   | `/checkout`          | Process checkout & create an order |
| POST   | `/complete/checkout` | complete the current pending checkout after receving payment refrence code and after verifying the payment from the provider |

### 🚚 Orders & Shipping
| Method | Endpoint               | Description |
|--------|----------------------|-------------|
| GET    | `/orders`            | Get user orders |
| GET    | `/orders/:id`        | Get order details |
| POST   | `/orders/cancel/:id` | Cancel an order |
| GET    | `/shipping/:orderId` | Get shipping status |

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## This is a personal project for learning purpose and implementing all services of my own, for now going with monolithic approach, ik microservices approches too

Note - some endpoints are yet to be implemented  
