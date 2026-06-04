import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { useAuth } from "./AuthProvider";

const WebSocketContext = createContext(null);

function getWebsocketUrl() {
  const backendUrl = import.meta.env.VITE_BACKEND_URL || window.location.origin;
  const baseUrl = new URL(backendUrl, window.location.origin);

  if (window.location.protocol === "https:" && baseUrl.protocol === "http:") {
    baseUrl.protocol = "wss:";
    baseUrl.host = window.location.host;
  }

  if (baseUrl.protocol === "https:") {
    baseUrl.protocol = "wss:";
  } else if (baseUrl.protocol === "http:") {
    baseUrl.protocol = "ws:";
  }
  baseUrl.pathname = "/backend/api/accounts/ws";
  baseUrl.search = "";
  return baseUrl.toString();
}

export function WebSocketProvider({ children }) {
  const { isLoggedIn, isLoading } = useAuth();
  const socketRef = useRef(null);
  const listenersRef = useRef({});
  const [status, setStatus] = useState("disconnected");

  useEffect(() => {
    if (isLoading || !isLoggedIn) {
      socketRef.current?.close();
      socketRef.current = null;
      setStatus("disconnected");
      return;
    }

    let reconnectTimeoutID;
    let shouldReconnect = true;

    const connect = () => {
      const socket = new WebSocket(getWebsocketUrl());
      socketRef.current = socket;
      setStatus("connecting");

      socket.addEventListener("open", () => {
        setStatus("connected");
      });

      socket.addEventListener("close", () => {
        if (socketRef.current === socket) {
          socketRef.current = null;
        }
        setStatus("disconnected");
        if (shouldReconnect) {
          reconnectTimeoutID = window.setTimeout(connect, 3000);
        }
      });

      socket.addEventListener("error", () => {
        setStatus("error");
      });

      socket.addEventListener("message", (event) => {
        let payload;
        try {
          payload = JSON.parse(event.data);
        } catch (parseError) {
          console.log("WEBSOCKET PAYLOAD ERROR", parseError);
          return;
        }

        const handlers = listenersRef.current[payload.type] || [];
        handlers.forEach((handler) => handler(payload));
      });
    };

    connect();

    return () => {
      shouldReconnect = false;
      window.clearTimeout(reconnectTimeoutID);
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [isLoading, isLoggedIn]);

  const send = useCallback((payload) => {
    if (socketRef.current?.readyState !== WebSocket.OPEN) {
      return false;
    }

    socketRef.current.send(JSON.stringify(payload));
    return true;
  }, []);

  const subscribe = useCallback((type, handler) => {
    if (!listenersRef.current[type]) {
      listenersRef.current[type] = [];
    }

    listenersRef.current[type].push(handler);

    return () => {
      listenersRef.current[type] = listenersRef.current[type].filter(
        (currentHandler) => currentHandler !== handler
      );
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ send, subscribe, status }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);

  if (!context) {
    throw new Error("useWebSocket must be used inside WebSocketProvider");
  }

  return context;
}
