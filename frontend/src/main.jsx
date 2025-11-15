import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Buffer } from "buffer";
import { AuthProvider } from "./contexts/AuthContext.jsx";
import App from "./App.jsx";
import "./styles/global.css";

// Polyfill Buffer for browser environment
window.Buffer = Buffer;
globalThis.Buffer = Buffer;

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <App />
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>
);
