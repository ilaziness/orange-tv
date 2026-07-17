import { BrowserRouter } from 'react-router'
import { AppRoutes } from './routes.tsx'
import './App.css'

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
