import { useEffect, useRef, useState } from "react";

const WS_URL =
  import.meta.env.VITE_WS_URL || "ws://localhost:8080";

export default function useWebSocket(pollId, onUpdate) {
  const socketRef = useRef(null);
  const [status, setStatus] = useState("connecting");

  useEffect(() => {
    if (!pollId) return;

    const socket = new WebSocket(
      `${WS_URL}/ws/poll/${pollId}`
    );

    socketRef.current = socket;
    setStatus("connecting");

    socket.onopen = () => {
      setStatus("connected");
    };

    socket.onmessage = (event) => {
      try {
        const update = JSON.parse(event.data);

        if (update?.counts) {
          onUpdate?.(update);
        }
      } catch (error) {
        console.error(
          "Invalid WebSocket message:",
          error
        );
      }
    };

    socket.onerror = () => {
      setStatus("error");
    };

    socket.onclose = () => {
      setStatus("disconnected");
    };

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [pollId, onUpdate]);

  return {
    status,
  };
}