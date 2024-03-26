import React from "react";
import { DisplayMessage } from "@/models/message";
import messageComponentStyle from "./MessageComponent.module.css";
import Discord from "../../../../public/discord-icon.svg";
import Twitch from "../../../../public/twitch-icon.svg";
import Vercel from "../../../../public/vercel-icon.svg";
import { Icon } from "@blueprintjs/core";

export default function MessageComponent({
  display,
}: {
  display: DisplayMessage;
}) {
  if (!display.message) return null;

  const attachments = display.message.attachments.map((attachment, index) => (
    <span
      key={index}
      className={`${messageComponentStyle.attachment} badge rounded-pill bg-secondary`}
    >
      <i className="bi bi-paperclip"></i>
      {attachment}
    </span>
  ));

  const getIcon = () => {
    return display.message.author.isBot || display.message.author.isAdmin ? (
      display.message.author.isBot ? (
        <Icon icon="cog"></Icon>
      ) : (
        <Icon icon="shield"></Icon>
      )
    ) : (
      <Icon icon="user" style={{ margin: "5px" }}></Icon>
    );
  };

  const content = display.message.messageParts.map((msgPart, index) => {
    if (msgPart.format) {
      return (
        <span key={index} style={{ color: msgPart.format.color }}>
          <b>{msgPart.content}</b>
        </span>
      );
    }

    if (msgPart.emoji) {
      return (
        <img
          key={index}
          alt={msgPart.emoji.alt}
          src={msgPart.emoji.id}
          className={messageComponentStyle.emoji}
        />
      );
    }

    return <span key={index}>{msgPart.content}</span>;
  });

  return (
    <div
      className={`grid grid-cols-12 gap-2 ${messageComponentStyle.messageObject}`}
    >
      <div className="col-span-1">
        {display.message.source == "discord" && (
          <img
            alt={"d"}
            src={Discord.src}
            className={messageComponentStyle.icon}
          />
        )}
        {display.message.source == "twitch" && (
          <img
            alt={"d"}
            src={Twitch.src}
            className={messageComponentStyle.icon}
          />
        )}
        {display.message.source == "test_source" && (
          <img
            alt={"d"}
            src={Vercel.src}
            className={messageComponentStyle.icon}
          />
        )}
      </div>
      <div className={`col-span-1`}>{getIcon()}</div>
      <div className={`col-span-3 ${messageComponentStyle.author}`}>
        <span style={{ color: display.message.author.color }}>
          <b>{display.message.author.username}</b>
        </span>
      </div>
      <div className="col-span-7">
        {!display.delete && (
          <>
            {/*<div className={messageComponentStyle.attachmentContainer}>*/}
            {/*  {attachments}*/}
            {/*</div>*/}
            <div className={messageComponentStyle.message}>
              {display.edit && <i>(edited)</i>}
              {content}
            </div>
          </>
        )}
        {display.delete && (
          <div className={messageComponentStyle.message}>
            <i>(deleted)</i>
          </div>
        )}
      </div>
    </div>
  );
}
