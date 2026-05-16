import { expect, test } from "@playwright/test";

test("login, tema, lead e persistência de sessão", async ({ page, context }) => {
  await context.clearCookies();
  await page.goto("/");

  await expect(page.getByRole("heading", { name: "Acesse sua conta" })).toBeVisible();

  const loginThemeSelect = page.getByTestId("theme-toggle");
  await loginThemeSelect.selectOption("light");
  await expect(page.locator("body")).toHaveAttribute("data-theme", "light");

  await page.getByPlaceholder("admin@juriscan.local").fill("admin@juriscan.local");
  await page.getByTestId("login-form").getByRole("button", { name: "Enviar código" }).click();

  const devTokenValue = page.locator("[data-testid='dev-token-hint'] strong");
  await expect(devTokenValue).toHaveText(/\d{6}/);
  const otpCode = (await devTokenValue.textContent())?.trim() || "";

  await page.getByTestId("otp-form").locator("input[placeholder='123456']").fill(otpCode);
  await page.getByTestId("otp-form").getByRole("button", { name: "Entrar" }).click();

  await expect(page.getByTestId("page-dashboard")).toBeVisible();
  await expect(page.locator("body")).toHaveAttribute("data-theme", "light");
  await expect(page.locator(".user-chip")).toContainText("admin@juriscan.local");

  const dashboardThemeSelect = page.getByTestId("theme-toggle");
  await dashboardThemeSelect.selectOption("dark");
  await expect(page.locator("body")).toHaveAttribute("data-theme", "dark");

  const leadName = `Lead E2E ${Date.now()}`;
  await page.getByTestId("open-create-lead").click();

  const createForm = page.getByTestId("create-lead-form");
  await createForm.getByLabel("Nome").fill(leadName);
  await createForm.getByLabel("Telefone").fill("(11) 99999-9999");
  await createForm.getByLabel("Origem").fill("whatsapp");
  await createForm.getByLabel("Assunto").fill("Teste de fluxo E2E");
  await createForm.getByLabel("Responsável").fill("admin@juriscan.local");
  await createForm.getByRole("button", { name: "Salvar lead" }).click();

  const createdLeadCard = page.getByTestId("lead-item").filter({ hasText: leadName }).first();
  await expect(createdLeadCard).toBeVisible();

  await createdLeadCard.getByRole("button", { name: "Editar" }).click();
  const editForm = page.getByTestId("edit-lead-form");
  await editForm.getByLabel("Assunto").fill("Teste editado no E2E");
  await editForm.getByLabel("Etapa").selectOption("qualificado");
  await editForm.getByRole("button", { name: "Salvar edição" }).click();

  await expect(page.getByTestId("column-qualificado")).toContainText(leadName);

  await page.reload();
  await expect(page.getByTestId("page-dashboard")).toBeVisible();
  await expect(page.locator(".user-chip")).toContainText("admin@juriscan.local");
  await expect(page.getByTestId("column-qualificado")).toContainText(leadName);

  await page.getByRole("button", { name: "Sair" }).click();
  await expect(page.getByRole("heading", { name: "Acesse sua conta" })).toBeVisible();
});
