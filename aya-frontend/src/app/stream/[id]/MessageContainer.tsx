import React from "react";
import MessageComponent from "./MessageComponent";
import messageContainerStyle from "./MessageContainer.module.css";
import { DisplayMessage } from "@/models/message";

export default function MessageContainer(props: {
  width: number;
  height: number;
  displayMsgs: DisplayMessage[];
}) {
  const getColor = (r: number, g: number, b: number): string => {
    return `rgba(${r}, ${g}, ${b}, ${0})`;
  };
  const backgroundColor = getColor(44, 47, 51);
  // const inactiveBackgroundColor = getColor(35, 39, 42);

  return (
    <div
      className={messageContainerStyle.display}
      style={{
        width: `${props.width}px`,
        height: `${props.height}px`,
        maxWidth: `${props.width}px`,
        maxHeight: `${props.height}px`,
        backgroundColor: "rgb(100,100,100,0.2)",
      }}
    >
      <div className={`container ${messageContainerStyle.messageContainer}`}>
        {props.displayMsgs.map((displayMsg, index) => (
          <MessageComponent key={index} display={displayMsg} />
        ))}
      </div>
    </div>
  );
}
