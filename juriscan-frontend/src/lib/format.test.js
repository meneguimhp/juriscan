import { describe, expect, it } from "vitest";
import { formatPhoneBR, onlyDigits } from "./format";

describe("format helpers", () => {
  it("onlyDigits keeps only numbers", () => {
    expect(onlyDigits("(11) 99999-9999")).toBe("11999999999");
  });

  it("formatPhoneBR formats BR mobile number", () => {
    expect(formatPhoneBR("11999999999")).toBe("(11) 99999-9999");
  });

  it("formatPhoneBR formats BR landline number", () => {
    expect(formatPhoneBR("1133334444")).toBe("(11) 3333-4444");
  });
});

