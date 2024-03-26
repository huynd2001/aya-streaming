"use client";

import { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import { DisplayMessage, MessageUpdate } from "@/models/message";
import MessageContainer from "@/app/stream/[id]/MessageContainer";

export default function Chat({ params }: { params: { id: string } }) {
  const wsURL = `${process.env.NEXT_PUBLIC_AYA_WEBSOCKET as string}/stream/${
    params.id
  }`;

  const [displayMsgs, setDisplayMsgs] = useState<DisplayMessage[]>([]);

  const [socketUrl, setSocketUrl] = useState(wsURL);

  const { sendMessage, lastJsonMessage, readyState } =
    useWebSocket<MessageUpdate>(socketUrl);

  const connectionStatus = {
    [ReadyState.CONNECTING]: "Connecting",
    [ReadyState.OPEN]: "Open",
    [ReadyState.CLOSING]: "Closing",
    [ReadyState.CLOSED]: "Closed",
    [ReadyState.UNINSTANTIATED]: "Uninstantiated",
  }[readyState];

  useEffect(() => {
    if (lastJsonMessage !== null) {
      console.log(lastJsonMessage);
      switch (lastJsonMessage.update) {
        case "new":
          setDisplayMsgs((msgs) => {
            console.log(msgs);
            return [
              ...msgs,
              {
                message: lastJsonMessage.message,
              },
            ];
          });
      }
    }
  }, [lastJsonMessage]);

  return (
    <div>
      <div>Connection Status: {connectionStatus}</div>
      <div>Open console log lmao</div>
      <MessageContainer width={400} height={600} displayMsgs={displayMsgs} />
    </div>
  );
}
