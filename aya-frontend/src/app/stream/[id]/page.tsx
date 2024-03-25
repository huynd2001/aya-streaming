"use client";

import { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";

export default function Chat({ params }: { params: { id: string } }) {
  const wsURL = `${process.env.NEXT_PUBLIC_AYA_WEBSOCKET as string}/stream/${
    params.id
  }`;

  const [socketUrl, setSocketUrl] = useState(wsURL);

  const { sendMessage, lastMessage, readyState } = useWebSocket(socketUrl);

  const connectionStatus = {
    [ReadyState.CONNECTING]: "Connecting",
    [ReadyState.OPEN]: "Open",
    [ReadyState.CLOSING]: "Closing",
    [ReadyState.CLOSED]: "Closed",
    [ReadyState.UNINSTANTIATED]: "Uninstantiated",
  }[readyState];

  useEffect(() => {
    if (lastMessage !== null) {
      console.log(lastMessage.data);
    }
  }, [lastMessage]);

  return (
    <div>
      <div>Connection Status: {connectionStatus}</div>
      <div>Open console log lmao</div>
    </div>
  );
}
