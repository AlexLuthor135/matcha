import { createRoot } from 'react-dom/client'
import App from './App.jsx'
import AuthProvider from './jsx/AuthProvider.jsx'
import { WebSocketProvider } from './jsx/WebSocketProvider.jsx'
import './css/index.css'

createRoot(document.getElementById('root')).render(
    <AuthProvider>
      <WebSocketProvider>
        <App />
      </WebSocketProvider>
    </AuthProvider>
)
