"use client";

import { Alignment, Button, Navbar } from "@blueprintjs/core";

import "normalize.css/normalize.css";
import "@blueprintjs/core/lib/css/blueprint.css";
import "@blueprintjs/icons/lib/css/blueprint-icons.css";
import { User } from "oidc-client-ts";
import { AuthProvider, useAuth } from "react-oidc-context";
import Image from "next/image";
import { useEffect, useState } from "react";

import { sha256 } from "js-sha256";

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

function ProfileModal(props: { person: UserProfile }) {
  return (
    <Button className="bp5-minimal">
      <Image
        alt={"ava"}
        src={props.person.avatar_url || "next.svg"}
        width={"20"}
        height={"20"}
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
    <Navbar>
      <Navbar.Group align={Alignment.LEFT}>
        <Navbar.Heading>Aya~</Navbar.Heading>
      </Navbar.Group>
      <Navbar.Group align={Alignment.RIGHT}>
        {user ? (
          <ProfileModal person={user}></ProfileModal>
        ) : (
          <Button
            className="bp5-minimal"
            icon="user"
            onClick={() => void auth.signinRedirect()}
          />
        )}
      </Navbar.Group>
    </Navbar>
  );
}

export default function Page() {
  return (
    <AuthProvider {...oidcConfig}>
      <AyaNavBar></AyaNavBar>
    </AuthProvider>
  );
}
