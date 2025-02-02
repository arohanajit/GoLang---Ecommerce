# E-Commerce Platform

A modern, full-stack e-commerce platform built with React, TypeScript, and Chakra UI for the frontend, with a microservices backend.

## Features

- 🛍️ **Product Management**
  - Browse products with search and category filters
  - Detailed product pages with images and descriptions
  - Real-time stock tracking
  - Price display and product categorization

- 🛒 **Shopping Cart**
  - Add/remove products
  - Adjust quantities
  - Persistent cart storage
  - Quick checkout process

- 👤 **User Management**
  - User registration and authentication
  - Profile management
  - Order history tracking
  - Secure password handling

- 📱 **Responsive Design**
  - Mobile-first approach
  - Adaptive layouts
  - Consistent experience across devices

## Tech Stack

### Frontend
- **React 18** - UI library
- **TypeScript** - Type safety and better developer experience
- **Chakra UI** - Component library for consistent design
- **React Router** - Client-side routing
- **React Query** - Server state management
- **Axios** - HTTP client
- **Vite** - Build tool and development server

### Backend (Microservices)
- Product Service
- User Service
- Order Service
- Authentication Service

## Getting Started

### Prerequisites
- Node.js (v18 or later)
- npm or yarn
- Go (for backend services)

### Installation

1. Clone the repository:
\`\`\`bash
git clone [repository-url]
cd e-commerce-platform
\`\`\`

2. Install frontend dependencies:
\`\`\`bash
cd frontend
npm install
\`\`\`

3. Start the development server:
\`\`\`bash
npm run dev
\`\`\`

The application will be available at http://localhost:5173

## Project Structure

\`\`\`
frontend/
├── src/
│   ├── api/          # API client and configurations
│   ├── components/   # Reusable UI components
│   ├── pages/        # Page components
│   ├── types/        # TypeScript interfaces
│   ├── theme.ts      # Chakra UI theme configuration
│   └── App.tsx       # Root component
├── public/           # Static assets
└── package.json      # Dependencies and scripts
\`\`\`

## Available Scripts

- \`npm run dev\` - Start development server
- \`npm run build\` - Build for production
- \`npm run lint\` - Run ESLint
- \`npm run preview\` - Preview production build

## API Integration

The frontend connects to a backend API running at \`http://localhost:3000/api\`. Key endpoints:

- \`/auth\` - Authentication endpoints (login/register)
- \`/products\` - Product management
- \`/users\` - User profile management
- \`/orders\` - Order processing and history

## Contributing

1. Fork the repository
2. Create your feature branch (\`git checkout -b feature/amazing-feature\`)
3. Commit your changes (\`git commit -m 'Add some amazing feature'\`)
4. Push to the branch (\`git push origin feature/amazing-feature\`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Chakra UI for the component library
- React Query for data fetching
- The open-source community for inspiration and tools
