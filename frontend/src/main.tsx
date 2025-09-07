import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

// Render the main application component into the root DOM element
createRoot(document.getElementById("root")!).render(<App />);
