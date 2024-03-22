"use client";

import { Alignment, Button, Navbar } from "@blueprintjs/core";

import { useSearchParams } from "next/navigation";

import "normalize.css/normalize.css";
import "@blueprintjs/core/lib/css/blueprint.css";
import "@blueprintjs/icons/lib/css/blueprint-icons.css";
import { useEffect } from "react";

export default function Page() {
  const searchParams = useSearchParams();
  // console.log(params);
  useEffect(() => {
    const id_token = searchParams.get("id_token");
    console.log(id_token);
  }, [searchParams]);
  return (
    <Navbar>
      <Navbar.Group align={Alignment.LEFT}>
        <Navbar.Heading>Aya~</Navbar.Heading>
      </Navbar.Group>
      <Navbar.Group align={Alignment.RIGHT}>
        <Button className="bp5-minimal" icon="user" text="Log in" />
      </Navbar.Group>
    </Navbar>
  );
}
