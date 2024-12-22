# MobileHub Backend API

MobileHub is a robust backend API for an e-commerce platform, developed using Go, the Gin framework, and PostgreSQL as the database. This project implements essential features for both admins and users, providing seamless management and interaction capabilities.

## Features

### Admin Features
- **Product Management**: Create, edit, and delete products.
- **User Management**: Block or unblock users.
- **Order Management**: List and change the status of orders.
- **Sales Reports**: View sales reports (daily, monthly, yearly).

### User Features
- **Product Browsing**: View available products.
- **Cart and Wishlist**: Add products to cart or wishlist.
- **Order Management**: Place or cancel orders using:
  - Cash on Delivery (COD)
  - Online Payments (integrated with Razorpay)
  - Wallet
- **Invoice Generation**: Generate invoices for orders.

## Technical Highlights
- **Authentication/Authorization**: Implemented using JWT (JSON Web Tokens).
- **Database**: PostgreSQL for reliable data storage and management.
- **Deployment**: Deployed on AWS EC2. Created a Dockerfile and deployed the application on DockerHub.
