import React from "react";

type Props = {
  type?: string;
  autoComplete?: string;
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
};

export function Input({
  type = "text",
  autoComplete = "",
  placeholder,
  value,
  onChange,
}: Props): React.ReactElement {
  return (
    <input
      type={type}
      placeholder={placeholder}
      autoComplete={autoComplete}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="block w-full rounded-md border-0 p-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-inset focus:ring-black sm:text-sm sm:leading-6"
    />
  );
}
