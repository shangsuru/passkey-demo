import React from "react";
import { Divider } from "../components/layout/Divider";
import { Link } from "../components/input/Link";
import { Layout } from "../components/layout/Layout";
import { PasswordSignUp } from "../components/form/PasswordSignUp";
import { PasskeySignUp } from "../components/form/PasskeySignUp";

export default function Register(): React.ReactElement {
  return (
    <Layout title="Sign up">
      <PasswordSignUp />
      <Divider />
      <PasskeySignUp />

      <div className="mt-10">
        <Link href="/" linkText="Already have an account?" />
      </div>
    </Layout>
  );
}
