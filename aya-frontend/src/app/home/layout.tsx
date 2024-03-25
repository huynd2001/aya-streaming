"use client";

import React, { useEffect, useState } from "react";
import { Inter } from "next/font/google";
import {
  Alignment,
  Button,
  Icon,
  IconSize,
  Menu,
  Navbar,
  Popover,
} from "@blueprintjs/core";

import "normalize.css/normalize.css";
import "@blueprintjs/core/lib/css/blueprint.css";
import "@blueprintjs/icons/lib/css/blueprint-icons.css";
import { User } from "oidc-client-ts";
import { AuthProvider, useAuth } from "react-oidc-context";
import Image from "next/image";

import { sha256 } from "js-sha256";

const inter = Inter({ subsets: ["latin"] });

const oidcConfig = {
  authority: process.env.NEXT_PUBLIC_AUTHORITY as string,
  client_id: process.env.NEXT_PUBLIC_CLIENT_ID as string,
  redirect_uri: process.env.NEXT_PUBLIC_REDIRECT_URL as string,
  response_type: "code",
  scope: "openid profile email",
  onSigninCallback: (_user: User | void): void => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
};

interface UserProfile {
  username?: string;
  avatar_url?: string;
}

function LoginPopOver({
  button,
  signInButton,
  signOutButton,
  isAuthenticated,
}: {
  button: Readonly<React.ReactNode>;
  signInButton: Readonly<React.ReactNode>;
  signOutButton: Readonly<React.ReactNode>;
  isAuthenticated: boolean;
}) {
  return (
    <Popover
      interactionKind="click"
      content={<Menu>{isAuthenticated ? signOutButton : signInButton}</Menu>}
      placement="auto-end"
      renderTarget={({ isOpen, ...targetProps }) => (
        <div {...targetProps}>{button}</div>
      )}
    ></Popover>
  );
}

function ProfileButton(props: { person: UserProfile }) {
  return (
    <Button className="bp5-minimal">
      <Image
        alt={"ava"}
        src={props.person.avatar_url || "next.svg"}
        width={IconSize.LARGE}
        height={IconSize.LARGE}
      ></Image>
    </Button>
  );
}

function AyaNavBar() {
  const auth = useAuth();
  const [user, setUser] = useState<UserProfile | undefined>(undefined);
  useEffect(() => {
    if (auth.isAuthenticated) {
      const address = String(auth.user?.profile.email).trim().toLowerCase();
      const email_hash = sha256(address);
      const ava_url = `https://www.gravatar.com/avatar/${email_hash}`;
      setUser({
        username: auth.user?.profile.preferred_username || "unknown",
        avatar_url: ava_url,
      });
    }
  }, [auth]);

  if (auth.error) {
    return (
      <div>
        Error occured! Please try refresh the web page or contact our on call
        team
      </div>
    );
  }
  return (
    <Navbar className={auth.isLoading ? "bp5-skeleton" : ""}>
      <Navbar.Group align={Alignment.LEFT}>
        <Navbar.Heading>Aya~</Navbar.Heading>
      </Navbar.Group>
      <Navbar.Group align={Alignment.RIGHT}>
        <LoginPopOver
          button={
            user ? (
              <ProfileButton person={user}></ProfileButton>
            ) : (
              <Button className="bp5-minimal">
                <Icon icon="user" size={IconSize.LARGE}></Icon>
              </Button>
            )
          }
          signInButton={
            <Button
              className="bp5-minimal"
              text="Sign In"
              onClick={() => void auth.signinRedirect()}
            ></Button>
          }
          signOutButton={
            <Button
              className="bp5-minimal"
              text="Sign Out"
              onClick={() => void auth.signoutRedirect()}
            ></Button>
          }
          isAuthenticated={auth.isAuthenticated}
        ></LoginPopOver>
      </Navbar.Group>
    </Navbar>
  );
}

function Page({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <AuthProvider {...oidcConfig}>
      <AyaNavBar></AyaNavBar>
      <div className={"main-page-wrapper"}>{children}</div>
    </AuthProvider>
  );
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Page>{children}</Page>
      </body>
    </html>
  );
}
