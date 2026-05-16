import { expect, test } from "@playwright/test";

function toDateTimeLocalPlus(hoursAhead = 2) {
  const date = new Date(Date.now() + hoursAhead * 60 * 60 * 1000);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

async function loginAsAdmin(page, context) {
  await context.clearCookies();
  await page.goto("/");

  await expect(page.getByRole("heading", { name: "Acesse sua conta" })).toBeVisible();
  await page.getByPlaceholder("admin@juriscan.local").fill("admin@juriscan.local");
  await page.getByTestId("login-form").getByRole("button", { name: "Enviar código" }).click();

  const devTokenValue = page.locator("[data-testid='dev-token-hint'] strong");
  await expect(devTokenValue).toHaveText(/\d{6}/);
  const otpCode = (await devTokenValue.textContent())?.trim() || "";

  await page.getByTestId("otp-form").locator("input[placeholder='123456']").fill(otpCode);
  await page.getByTestId("otp-form").getByRole("button", { name: "Entrar" }).click();

  await expect(page.getByTestId("page-dashboard")).toBeVisible();
}

test("sprint 3-6: triagem IA, follow-up, publicações, prazos e compliance", async ({ page, context }) => {
  await loginAsAdmin(page, context);

  const suffix = Date.now();
  const leadName = `Lead S36 ${suffix}`;
  const sourceName = `DJE-E2E-${suffix}`;
  const templateName = `Template S36 ${suffix}`;

  await page.getByTestId("open-create-lead").click();
  const createForm = page.getByTestId("create-lead-form");
  await createForm.getByLabel("Nome").fill(leadName);
  await createForm.getByLabel("Telefone").fill("(11) 97777-7711");
  await createForm.getByLabel("Origem").fill("whatsapp");
  await createForm.getByLabel("Assunto").fill("Caso trabalhista com audiencia marcada");
  await createForm.getByLabel("Responsável").fill("admin@juriscan.local");
  await createForm.getByRole("button", { name: "Salvar lead" }).click();

  const leadCard = page.getByTestId("lead-item").filter({ hasText: leadName }).first();
  await expect(leadCard).toBeVisible();
  const triageResponsePromise = page.waitForResponse(
    (resp) =>
      resp.request().method() === "POST" &&
      resp.url().includes("/v1/crm/leads/") &&
      resp.url().includes("/triage")
  );
  await leadCard.getByRole("button", { name: "Triar IA" }).click();
  const triageResponse = await triageResponsePromise;
  if (triageResponse.status() >= 400) {
    throw new Error(`triage failed: ${triageResponse.status()} - ${await triageResponse.text()}`);
  }
  await expect(leadCard).toContainText("IA:");

  const applySuggestion = leadCard.getByRole("button", { name: "Aplicar sugestão" });
  if (await applySuggestion.isVisible()) {
    await applySuggestion.click();
    await expect(leadCard).toContainText("Próximo passo:");
  }

  await page.getByTestId("nav-followups").click();
  await expect(page.getByTestId("page-followups")).toBeVisible();

  const templateForm = page.locator("form").filter({ hasText: "Novo template de resposta" }).first();
  await templateForm.getByLabel("Nome").fill(templateName);
  await templateForm.getByLabel("Mensagem").fill("Retorno padrao para contato comercial.");
  await templateForm.getByRole("button", { name: "Salvar template" }).click();
  await expect(page.locator(".panel-sub").filter({ hasText: "Templates cadastrados" })).toContainText(templateName);

  const followupForm = page.locator("form").filter({ hasText: "Agendar follow-up" }).first();
  await followupForm.getByLabel("Lead").selectOption({ label: leadName });
  await followupForm.getByLabel("Template (opcional)").selectOption({ label: templateName });
  await followupForm.getByLabel("Mensagem").fill("Reforcar envio de documentos.");
  await followupForm.getByLabel("Data e hora do follow-up").fill(toDateTimeLocalPlus(3));
  await followupForm.getByRole("button", { name: "Agendar follow-up" }).click();

  const followupItem = page
    .locator(".panel-sub")
    .filter({ hasText: "Follow-ups" })
    .locator("li")
    .filter({ hasText: leadName })
    .first();
  await expect(followupItem).toBeVisible();
  if (await followupItem.getByRole("button", { name: "Concluir" }).isVisible()) {
    await followupItem.getByRole("button", { name: "Concluir" }).click();
    await expect(followupItem).toContainText("Concluído");
  }

  await page.getByTestId("nav-publications").click();
  await expect(page.getByTestId("page-publications")).toBeVisible();

  const publicationForm = page.locator("form").filter({ hasText: "Registrar publicação" }).first();
  await publicationForm.getByLabel("Origem").fill(sourceName);
  await publicationForm.getByLabel("Conteúdo").fill(
    "Despacho: prazo de 5 dias uteis para manifestacao. Intimacao da parte autora."
  );
  await publicationForm.getByRole("button", { name: "Salvar publicação" }).click();

  const publicationItem = page.locator(".conversation-item").filter({ hasText: sourceName }).first();
  await expect(publicationItem).toBeVisible();
  await publicationItem.getByRole("button", { name: "Analisar IA" }).click();
  await expect(publicationItem).toContainText("Confiança IA:");

  await publicationItem.getByLabel("Prazo final (humano)").fill(toDateTimeLocalPlus(24));
  await publicationItem.getByLabel("Dono da tarefa").fill("admin@juriscan.local");
  await publicationItem.getByLabel("Observações").fill("Prazo validado para operação E2E.");
  await publicationItem.getByRole("button", { name: "Confirmar prazo e criar tarefa" }).click();
  await expect(publicationItem).toContainText("Prazo validado por");

  await page.getByTestId("nav-deadlines").click();
  await expect(page.getByTestId("page-deadlines")).toBeVisible();
  await expect(page.locator(".conversation-item").first()).toContainText("Responsável");
  const statusSelect = page.locator(".conversation-item").first().getByLabel("Status");
  await statusSelect.selectOption("em_execucao");
  await expect(statusSelect).toHaveValue("em_execucao");

  await page.getByTestId("nav-compliance").click();
  await expect(page.getByTestId("page-compliance")).toBeVisible();
  await expect(page.locator(".data-table")).toContainText("lead_triage");
  await expect(page.locator(".data-table")).toContainText("publication_analysis");
});
