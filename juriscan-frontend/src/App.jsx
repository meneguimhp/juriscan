import { useEffect, useMemo, useState } from "react";
import { formatPhoneBR, onlyDigits } from "./lib/format";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
const THEME_PREF_KEY = "juriscan.theme.pref";
const BRAND_ICON = "/logo-juriscan-icon-card.png?v=20260514-1";
const BRAND_LOGO_DARK = "/logo-juriscan-dark-smooth.png?v=20260414-15";

const STAGES = ["novo", "triado", "qualificado", "proposta", "fechado", "perdido"];
const STAGE_LABEL = {
  novo: "Novo",
  triado: "Triado",
  qualificado: "Qualificado",
  proposta: "Proposta",
  fechado: "Fechado",
  perdido: "Perdido"
};

const CONVERSATION_STATUS_LABEL = {
  nova: "Nova",
  sem_lead: "Sem lead",
  vinculada: "Vinculada"
};

const USER_ROLES = [
  { value: "admin", label: "Administrador" },
  { value: "controller", label: "Controladoria" },
  { value: "lawyer", label: "Advogado" },
  { value: "commercial", label: "Comercial" }
];

const USER_ROLE_LABEL = Object.fromEntries(USER_ROLES.map((role) => [role.value, role.label]));
const USER_STATUS_LABEL = {
  active: "Ativo",
  inactive: "Inativo"
};

const FOLLOW_UP_STATUS_LABEL = {
  pendente: "Pendente",
  concluido: "Concluído",
  cancelado: "Cancelado"
};

const DEADLINE_STATUS_OPTIONS = [
  { value: "aberto", label: "Aberto" },
  { value: "em_execucao", label: "Em execução" },
  { value: "concluido", label: "Concluído" }
];

function loadThemePreference() {
  try {
    const stored = window.localStorage.getItem(THEME_PREF_KEY);
    if (stored === "light" || stored === "dark") return stored;
  } catch {}
  return "dark";
}

function saveThemePreference(value) {
  try {
    window.localStorage.setItem(THEME_PREF_KEY, value);
  } catch {}
}

async function apiFetch(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    ...options
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || "Erro inesperado na API.");
  return data;
}

function friendlyAuthError(message) {
  const normalized = String(message || "").toLowerCase();
  if (normalized.includes("invalid credentials") || normalized.includes("unauthorized")) {
    return "Este e-mail nao esta cadastrado no piloto. Peca para um administrador incluir em Usuarios.";
  }
  return message || "Nao foi possivel solicitar o codigo de acesso.";
}

function ThemeSwitch({ themePreference, onChange, compact = false }) {
  return (
    <label className={`theme-picker ${compact ? "theme-picker-compact" : ""}`}>
      <span className="theme-label">Tema</span>
      <select
        className="theme-select"
        data-testid="theme-toggle"
        aria-label="Tema"
        value={themePreference}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="dark">Escuro</option>
        <option value="light">Claro</option>
      </select>
    </label>
  );
}

function BrandLogo({ compact = false, dark = false }) {
  if (dark) {
    return <img className={compact ? "logo-topbar" : "logo-auth"} src={BRAND_LOGO_DARK} alt="Juriscan" />;
  }

  return (
    <div className={`brand-logo ${compact ? "brand-logo-compact" : ""}`} aria-label="Juriscan Inteligência Jurídica">
      <img className="brand-logo-icon" src={BRAND_ICON} alt="" aria-hidden="true" />
      <div className="brand-logo-copy">
        <div className="brand-logo-name">
          <span className="brand-logo-juri">JURI</span>
          <span className="brand-logo-scan">SCAN</span>
        </div>
        <div className="brand-logo-tagline">
          <span className="brand-logo-rule" />
          <span>INTELIGÊNCIA JURÍDICA</span>
          <span className="brand-logo-rule" />
        </div>
      </div>
    </div>
  );
}

function statusTagClass(status) {
  if (status === "estourado") return "tag tag-danger";
  if (status === "em_aberto") return "tag tag-warn";
  if (status === "atendido_fora_prazo") return "tag tag-muted";
  return "tag tag-success";
}

function resolveLeadSLAStatus(lead) {
  if (lead?.sla_status) return lead.sla_status;
  if (lead?.stage === "novo" || lead?.stage === "triado") return "em_aberto";
  return "atendido_no_prazo";
}

function resolveLeadMinutesWithoutResponse(lead) {
  if (Number.isFinite(lead?.minutes_without_response)) return lead.minutes_without_response;
  if (!lead?.created_at) return 0;
  const createdAt = new Date(lead.created_at);
  if (Number.isNaN(createdAt.getTime())) return 0;
  if (lead?.stage !== "novo" && lead?.stage !== "triado") return 0;
  const diffMs = Date.now() - createdAt.getTime();
  return Math.max(0, Math.floor(diffMs / 60000));
}

function resolveUserRoleLabel(role) {
  return USER_ROLE_LABEL[role] || role;
}

function resolveUserStatusLabel(status) {
  return USER_STATUS_LABEL[status] || status;
}

function WorkQueueCard({ label, title, detail, count, action, onClick, tone = "info" }) {
  return (
    <button type="button" className={`queue-card queue-card-${tone}`} onClick={onClick}>
      <span className="queue-label">{label}</span>
      <strong>{title}</strong>
      <span className="queue-detail">{detail}</span>
      <span className="queue-footer">
        <b>{count}</b>
        <span>{action}</span>
      </span>
    </button>
  );
}

function toDateTimeLocalInput(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const offsetDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return offsetDate.toISOString().slice(0, 16);
}

function fromDateTimeLocalInput(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString();
}

export function App() {
  const [themePreference, setThemePreference] = useState(loadThemePreference);
  const isDarkMode = themePreference === "dark";

  const [booting, setBooting] = useState(true);
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [otpRequested, setOtpRequested] = useState(false);
  const [devToken, setDevToken] = useState("");
  const [sessionToken, setSessionToken] = useState("");
  const [user, setUser] = useState(null);

  const [leads, setLeads] = useState([]);
  const [slaMinutes, setSlaMinutes] = useState(30);
  const [showOnlyOverdue, setShowOnlyOverdue] = useState(false);
  const [showLeadForm, setShowLeadForm] = useState(false);
  const [leadName, setLeadName] = useState("");
  const [leadPhone, setLeadPhone] = useState("");
  const [leadOrigin, setLeadOrigin] = useState("whatsapp");
  const [leadSubject, setLeadSubject] = useState("");
  const [leadOwnerEmail, setLeadOwnerEmail] = useState("");

  const [editingLeadId, setEditingLeadId] = useState("");
  const [editName, setEditName] = useState("");
  const [editPhone, setEditPhone] = useState("");
  const [editOrigin, setEditOrigin] = useState("whatsapp");
  const [editSubject, setEditSubject] = useState("");
  const [editOwnerEmail, setEditOwnerEmail] = useState("");
  const [editStage, setEditStage] = useState("novo");

  const [conversations, setConversations] = useState([]);
  const [conversationDrafts, setConversationDrafts] = useState({});
  const [simPhone, setSimPhone] = useState("");
  const [simName, setSimName] = useState("");
  const [simMessage, setSimMessage] = useState("");
  const [simCreateLead, setSimCreateLead] = useState(true);

  const [templates, setTemplates] = useState([]);
  const [followUps, setFollowUps] = useState([]);
  const [newTemplateName, setNewTemplateName] = useState("");
  const [newTemplateBody, setNewTemplateBody] = useState("");
  const [followUpLeadID, setFollowUpLeadID] = useState("");
  const [followUpTemplateID, setFollowUpTemplateID] = useState("");
  const [followUpMessage, setFollowUpMessage] = useState("");
  const [followUpDueAt, setFollowUpDueAt] = useState("");

  const [publications, setPublications] = useState([]);
  const [publicationSource, setPublicationSource] = useState("DJE");
  const [publicationInputType, setPublicationInputType] = useState("texto");
  const [publicationFileName, setPublicationFileName] = useState("");
  const [publicationRawText, setPublicationRawText] = useState("");
  const [publicationValidationDrafts, setPublicationValidationDrafts] = useState({});

  const [deadlineTasks, setDeadlineTasks] = useState([]);
  const [deadlineAlerts, setDeadlineAlerts] = useState([]);
  const [aiLogs, setAiLogs] = useState([]);

  const [users, setUsers] = useState([]);
  const [newUserName, setNewUserName] = useState("");
  const [newUserEmail, setNewUserEmail] = useState("");
  const [newUserRole, setNewUserRole] = useState("commercial");
  const [activePage, setActivePage] = useState("dashboard");

  const [error, setError] = useState("");
  const [status, setStatus] = useState("");

  function handleThemeChange(nextTheme) {
    setThemePreference(nextTheme);
    saveThemePreference(nextTheme);
  }

  useEffect(() => {
    document.body.dataset.theme = isDarkMode ? "dark" : "light";
  }, [isDarkMode]);

  useEffect(() => {
    if (activePage === "users" && user?.role !== "admin") {
      setActivePage("dashboard");
    }
  }, [activePage, user?.role]);

  useEffect(() => {
    if (activePage === "publications" && !["admin", "controller", "lawyer"].includes(user?.role || "")) {
      setActivePage("dashboard");
    }
    if (activePage === "deadlines" && !["admin", "controller", "lawyer"].includes(user?.role || "")) {
      setActivePage("dashboard");
    }
    if (activePage === "compliance" && !["admin", "controller"].includes(user?.role || "")) {
      setActivePage("dashboard");
    }
  }, [activePage, user?.role]);

  async function fetchLeads({ onlyOverdue = showOnlyOverdue } = {}) {
    const query = onlyOverdue ? "?sla=estourado" : "";
    const data = await apiFetch(`/v1/crm/leads${query}`, {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setLeads(data.items || []);
    if (Number.isFinite(data.sla_minutes)) setSlaMinutes(data.sla_minutes);
  }

  async function fetchConversations() {
    const data = await apiFetch("/v1/whatsapp/conversations", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    const items = data.items || [];
    setConversations(items);
    setConversationDrafts((current) => {
      const next = { ...current };
      for (const item of items) {
        if (!next[item.id]) {
          next[item.id] = {
            status: item.status,
            leadID: item.lead_id || item.suggested_lead_id || ""
          };
        } else if (!next[item.id].leadID && (item.lead_id || item.suggested_lead_id)) {
          next[item.id] = {
            ...next[item.id],
            leadID: item.lead_id || item.suggested_lead_id || ""
          };
        }
      }
      return next;
    });
  }

  async function fetchUsers() {
    const data = await apiFetch("/v1/admin/users", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setUsers(data.items || []);
  }

  async function fetchTemplates() {
    const data = await apiFetch("/v1/crm/templates", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setTemplates(data.items || []);
  }

  async function fetchFollowUps({ onlyPending = false } = {}) {
    const query = onlyPending ? "?pending=true" : "";
    const data = await apiFetch(`/v1/crm/followups${query}`, {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setFollowUps(data.items || []);
  }

  async function fetchPublications() {
    const data = await apiFetch("/v1/publications", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setPublications(data.items || []);
  }

  async function fetchDeadlineTasks() {
    const data = await apiFetch("/v1/deadlines/tasks", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setDeadlineTasks(data.items || []);
  }

  async function fetchDeadlineAlerts() {
    const data = await apiFetch("/v1/deadlines/alerts", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setDeadlineAlerts(data.items || []);
  }

  async function fetchAILogs() {
    const data = await apiFetch("/v1/compliance/ai-logs", {
      method: "GET",
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
    });
    setAiLogs(data.items || []);
  }

  async function refreshDashboardData(options = {}) {
    const onlyOverdue = options.onlyOverdue ?? showOnlyOverdue;
    const role = options.role || user?.role;
    const tasks = [fetchLeads({ onlyOverdue }), fetchConversations(), fetchTemplates(), fetchFollowUps()];
    if (role === "admin") {
      tasks.push(fetchUsers());
    }
    if (["admin", "controller", "lawyer"].includes(role)) {
      tasks.push(fetchPublications(), fetchDeadlineTasks(), fetchDeadlineAlerts());
    }
    if (["admin", "controller"].includes(role)) {
      tasks.push(fetchAILogs());
    }
    const results = await Promise.allSettled(tasks);
    const [leadResult, conversationResult] = results;
    if (leadResult.status === "rejected") throw leadResult.reason;
    if (conversationResult.status === "rejected") {
      setStatus("Sessão ativa. Módulo WhatsApp indisponível no backend atual.");
    }
  }

  useEffect(() => {
    let cancelled = false;

    async function restoreSession() {
      try {
        const data = await apiFetch("/v1/identity/me", { method: "GET" });
        if (cancelled) return;
        setUser(data.user);
        setLeadOwnerEmail(data.user?.email || "");
        await refreshDashboardData({ onlyOverdue: false, role: data.user?.role });
        setStatus("Sessão restaurada.");
      } catch {
        if (!cancelled) setUser(null);
      } finally {
        if (!cancelled) setBooting(false);
      }
    }

    restoreSession();
    return () => {
      cancelled = true;
    };
  }, []);

  const leadsByStage = useMemo(() => {
    const grouped = Object.fromEntries(STAGES.map((stage) => [stage, []]));
    for (const lead of leads) {
      const stage = STAGES.includes(lead.stage) ? lead.stage : "novo";
      grouped[stage].push(lead);
    }
    return grouped;
  }, [leads]);

  const openQueueItems = useMemo(() => {
    return leads.filter((lead) => {
      const status = resolveLeadSLAStatus(lead);
      return status === "em_aberto" || status === "estourado";
    });
  }, [leads]);
  const overdueCount = useMemo(
    () => openQueueItems.filter((lead) => resolveLeadSLAStatus(lead) === "estourado").length,
    [openQueueItems]
  );

  const leadStats = useMemo(
    () => ({
      total: leads.length,
      novos: leadsByStage.novo.length,
      qualificados: leadsByStage.qualificado.length,
      fechados: leadsByStage.fechado.length
    }),
    [leads.length, leadsByStage]
  );

  const pendingConversationsCount = useMemo(
    () => conversations.filter((item) => item.status === "nova" || item.status === "sem_lead").length,
    [conversations]
  );
  const pendingFollowUpsCount = useMemo(
    () => followUps.filter((item) => item.status === "pendente").length,
    [followUps]
  );
  const pendingPublicationsCount = useMemo(
    () => publications.filter((item) => !item.analysis).length,
    [publications]
  );
  const pendingValidationsCount = useMemo(
    () => publications.filter((item) => item.analysis && !item.validation).length,
    [publications]
  );
  const openDeadlineCount = useMemo(
    () => deadlineTasks.filter((item) => item.status !== "concluido").length,
    [deadlineTasks]
  );

  async function requestOTP(event) {
    event.preventDefault();
    setError("");
    setStatus("Solicitando código...");
    try {
      const data = await apiFetch("/v1/identity/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim() })
      });
      setOtpRequested(true);
      setDevToken(data.token || "");
      setStatus("Código enviado.");
    } catch (err) {
      setDevToken("");
      setError(friendlyAuthError(err.message));
      setStatus("Nao foi possivel enviar o codigo.");
    }
  }

  async function verifyOTP(event) {
    event.preventDefault();
    setError("");
    if (onlyDigits(otp).length !== 6) {
      setError("Informe um código OTP válido com 6 dígitos.");
      return;
    }
    setStatus("Validando código...");
    try {
      const data = await apiFetch("/v1/identity/auth/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim(), token: onlyDigits(otp) })
      });
      setSessionToken(data.access_token || "");
      setUser(data.user);
      setLeadOwnerEmail(data.user.email);
      setActivePage("dashboard");
      await refreshDashboardData({ onlyOverdue: false, role: data.user?.role });
      setStatus(`Autenticado como ${data.user.email}.`);
    } catch (err) {
      setError(err.message);
      setStatus("Falha na validação.");
    }
  }

  async function createLead(event) {
    event.preventDefault();
    if (!user) return;
    setError("");

    const phoneDigits = onlyDigits(leadPhone);
    if (phoneDigits.length < 10) {
      setError("Telefone inválido. Informe DDD + número.");
      return;
    }
    if (leadName.trim().length < 3) {
      setError("Nome do lead deve ter ao menos 3 caracteres.");
      return;
    }

    setStatus("Cadastrando lead...");
    try {
      await apiFetch("/v1/crm/leads", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          name: leadName.trim(),
          phone: phoneDigits,
          origin: leadOrigin.trim(),
          subject: leadSubject.trim(),
          owner_email: leadOwnerEmail.trim()
        })
      });
      setLeadName("");
      setLeadPhone("");
      setLeadSubject("");
      setShowLeadForm(false);
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      setStatus("Lead cadastrado com sucesso.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha no cadastro.");
    }
  }

  function startEditLead(lead) {
    setEditingLeadId(lead.id);
    setEditName(lead.name || "");
    setEditPhone(formatPhoneBR(lead.phone || ""));
    setEditOrigin(lead.origin || "whatsapp");
    setEditSubject(lead.subject || "");
    setEditOwnerEmail(lead.owner_email || "");
    setEditStage(lead.stage || "novo");
    setShowLeadForm(false);
    setError("");
  }

  function cancelEditLead() {
    setEditingLeadId("");
    setEditName("");
    setEditPhone("");
    setEditOrigin("whatsapp");
    setEditSubject("");
    setEditOwnerEmail("");
    setEditStage("novo");
  }

  async function updateLead(event) {
    event.preventDefault();
    if (!user || !editingLeadId) return;
    setError("");

    const phoneDigits = onlyDigits(editPhone);
    if (phoneDigits.length < 10) {
      setError("Telefone inválido. Informe DDD + número.");
      return;
    }
    if (editName.trim().length < 3) {
      setError("Nome do lead deve ter ao menos 3 caracteres.");
      return;
    }

    setStatus("Salvando edição...");
    try {
      await apiFetch(`/v1/crm/leads/${editingLeadId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          name: editName.trim(),
          phone: phoneDigits,
          origin: editOrigin.trim(),
          subject: editSubject.trim(),
          owner_email: editOwnerEmail.trim(),
          stage: editStage
        })
      });
      cancelEditLead();
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      setStatus("Lead atualizado com sucesso.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha na edição.");
    }
  }

  async function changeLeadStage(lead, stage) {
    if (!user || stage === lead.stage) return;
    setError("");
    try {
      await apiFetch(`/v1/crm/leads/${lead.id}/stage`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({ stage })
      });
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
    } catch (err) {
      setError(err.message);
    }
  }

  async function simulateWhatsAppMessage(event) {
    event.preventDefault();
    setError("");
    const phone = onlyDigits(simPhone);
    const contactName = simName.trim();
    const message = simMessage.trim();
    if (phone.length < 10) {
      setError("Telefone inválido para simulação de WhatsApp.");
      return;
    }
    if (message.length < 2) {
      setError("Mensagem muito curta para simulação.");
      return;
    }

    if (simCreateLead && contactName.length < 3) {
      setError("Informe o nome do contato para gerar o lead.");
      return;
    }

    setStatus(simCreateLead ? "Recebendo mensagem e gerando lead..." : "Recebendo mensagem na caixa WhatsApp...");
    try {
      const conversation = await apiFetch("/v1/whatsapp/simulate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          phone,
          message,
          contact_name: contactName
        })
      });
      if (simCreateLead) {
        const lead = await apiFetch("/v1/crm/leads", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
          },
          body: JSON.stringify({
            name: contactName,
            phone,
            origin: "whatsapp",
            subject: message,
            owner_email: user?.email || leadOwnerEmail.trim()
          })
        });
        await apiFetch(`/v1/whatsapp/conversations/${conversation.id}`, {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
            ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
          },
          body: JSON.stringify({
            status: "vinculada",
            lead_id: lead.id
          })
        });
        await fetchLeads({ onlyOverdue: showOnlyOverdue });
      }
      setSimPhone("");
      setSimName("");
      setSimMessage("");
      await fetchConversations();
      setStatus(simCreateLead ? "Mensagem recebida, lead criado e conversa vinculada." : "Mensagem recebida na caixa.");
    } catch (err) {
      const message =
        err.message === "Erro inesperado na API."
          ? "Webhook WhatsApp indisponível. Reinicie o backend para carregar as novas rotas da sprint."
          : err.message;
      setError(message);
      setStatus("Falha ao receber mensagem.");
    }
  }

  function updateConversationDraft(conversationId, field, value) {
    setConversationDrafts((current) => ({
      ...current,
      [conversationId]: {
        status: current[conversationId]?.status || "nova",
        leadID: current[conversationId]?.leadID || "",
        [field]: value
      }
    }));
  }

  async function saveConversationClassification(conversationId) {
    const draft = conversationDrafts[conversationId];
    if (!draft) return;
    setError("");
    try {
      await apiFetch(`/v1/whatsapp/conversations/${conversationId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          status: draft.status,
          lead_id: draft.leadID
        })
      });
      await fetchConversations();
      setStatus("Classificação da conversa atualizada.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao classificar conversa.");
    }
  }

  async function runLeadTriage(leadId) {
    setError("");
    setStatus("Executando triagem IA...");
    try {
      await apiFetch(`/v1/crm/leads/${leadId}/triage`, {
        method: "POST",
        headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
      });
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      if (["admin", "controller"].includes(user?.role)) await fetchAILogs();
      setStatus("Triagem IA concluída.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha na triagem IA.");
    }
  }

  async function applySuggestedNextStep(lead) {
    const nextStep = lead?.ai_classification?.suggested_action || lead?.next_step || "";
    if (!nextStep) return;
    setError("");
    setStatus("Aplicando próximo passo sugerido...");
    try {
      await apiFetch(`/v1/crm/leads/${lead.id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          next_step: nextStep
        })
      });
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      setStatus("Próximo passo aplicado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao aplicar próximo passo.");
    }
  }

  async function overrideLeadTriage(lead) {
    if (!lead?.ai_classification) return;
    const reason = window.prompt("Motivo do override humano:", "Ajuste de contexto do caso");
    if (!reason || !reason.trim()) return;

    const defaultCategory = lead.ai_classification.category || "consultivo";
    const defaultUrgency = lead.ai_classification.urgency || "media";
    const defaultScore = String(lead.ai_classification.score ?? 60);
    const defaultAction = lead.ai_classification.suggested_action || "agendar contato";
    const defaultJustification = lead.ai_classification.justification || "Ajuste humano.";

    const category = window.prompt("Categoria (ex: trabalhista, civel, penal):", defaultCategory) || defaultCategory;
    const urgency = window.prompt("Urgencia (baixa, media, alta):", defaultUrgency) || defaultUrgency;
    const scoreRaw = window.prompt("Score (0-100):", defaultScore) || defaultScore;
    const suggestedAction =
      window.prompt("Próximo passo sugerido:", defaultAction) || defaultAction;
    const justification =
      window.prompt("Justificativa:", defaultJustification) || defaultJustification;

    setError("");
    setStatus("Aplicando override humano...");
    try {
      await apiFetch(`/v1/crm/leads/${lead.id}/triage/override`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          reason: reason.trim(),
          category: category.trim(),
          urgency: urgency.trim(),
          score: Number.parseInt(scoreRaw, 10),
          suggested_action: suggestedAction.trim(),
          justification: justification.trim()
        })
      });
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      if (["admin", "controller"].includes(user?.role)) await fetchAILogs();
      setStatus("Override humano aplicado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha no override.");
    }
  }

  async function createTemplate(event) {
    event.preventDefault();
    if (!["admin", "commercial"].includes(user?.role)) return;
    setError("");
    try {
      await apiFetch("/v1/crm/templates", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          name: newTemplateName.trim(),
          channel: "whatsapp",
          body: newTemplateBody.trim()
        })
      });
      setNewTemplateName("");
      setNewTemplateBody("");
      await fetchTemplates();
      setStatus("Template cadastrado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao cadastrar template.");
    }
  }

  async function scheduleFollowUp(event) {
    event.preventDefault();
    setError("");
    if (!followUpLeadID) {
      setError("Selecione um lead para agendar o retorno.");
      return;
    }
    const dueAt = fromDateTimeLocalInput(followUpDueAt);
    if (!dueAt) {
      setError("Informe data/hora do retorno.");
      return;
    }
    let message = followUpMessage.trim();
    if (!message && followUpTemplateID) {
      const template = templates.find((item) => item.id === followUpTemplateID);
      message = template?.body || "";
    }
    if (!message) {
      setError("Mensagem do retorno obrigatoria.");
      return;
    }
    try {
      await apiFetch("/v1/crm/followups", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          lead_id: followUpLeadID,
          template_id: followUpTemplateID || "",
          message,
          due_at: dueAt
        })
      });
      setFollowUpLeadID("");
      setFollowUpTemplateID("");
      setFollowUpMessage("");
      setFollowUpDueAt("");
      await fetchFollowUps();
      await fetchLeads({ onlyOverdue: showOnlyOverdue });
      setStatus("Retorno agendado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao agendar retorno.");
    }
  }

  async function markFollowUpStatus(id, statusValue) {
    setError("");
    try {
      await apiFetch(`/v1/crm/followups/${id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({ status: statusValue })
      });
      await fetchFollowUps();
      setStatus("Status de retorno atualizado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao atualizar retorno.");
    }
  }

  async function createPublication(event) {
    event.preventDefault();
    setError("");
    if (!publicationRawText.trim()) {
      setError("Texto da publicação obrigatório.");
      return;
    }
    try {
      await apiFetch("/v1/publications", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          source: publicationSource.trim(),
          input_type: publicationInputType,
          file_name: publicationInputType === "arquivo" ? publicationFileName.trim() : "",
          raw_text: publicationRawText.trim()
        })
      });
      setPublicationRawText("");
      setPublicationFileName("");
      await fetchPublications();
      setStatus("Publicação registrada.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao registrar publicação.");
    }
  }

  async function analyzePublication(publicationId) {
    setError("");
    setStatus("Executando extracao assistida...");
    try {
      await apiFetch(`/v1/publications/${publicationId}/analyze`, {
        method: "POST",
        headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
      });
      await fetchPublications();
      if (["admin", "controller"].includes(user?.role)) await fetchAILogs();
      setStatus("Análise assistida concluída.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha na análise da publicação.");
    }
  }

  function updatePublicationValidationDraft(publicationId, field, value) {
    setPublicationValidationDrafts((current) => ({
      ...current,
      [publicationId]: {
        finalDeadlineAt: current[publicationId]?.finalDeadlineAt || "",
        notes: current[publicationId]?.notes || "",
        ownerEmail: current[publicationId]?.ownerEmail || "",
        [field]: value
      }
    }));
  }

  async function validatePublication(publicationId) {
    const draft = publicationValidationDrafts[publicationId];
    if (!draft) {
      setError("Informe o prazo final para validar.");
      return;
    }
    const deadlineISO = fromDateTimeLocalInput(draft.finalDeadlineAt);
    if (!deadlineISO) {
      setError("Prazo final inválido.");
      return;
    }
    setError("");
    setStatus("Validando prazo e gerando tarefa...");
    try {
      await apiFetch(`/v1/publications/${publicationId}/validate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          final_deadline_at: deadlineISO,
          notes: draft.notes || "",
          owner_email: draft.ownerEmail || ""
        })
      });
      await Promise.all([fetchPublications(), fetchDeadlineTasks(), fetchDeadlineAlerts()]);
      setStatus("Publicação validada e tarefa criada.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao validar publicação.");
    }
  }

  async function updateDeadlineTask(taskId, payload) {
    setError("");
    try {
      await apiFetch(`/v1/deadlines/tasks/${taskId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify(payload)
      });
      await Promise.all([fetchDeadlineTasks(), fetchDeadlineAlerts()]);
      setStatus("Tarefa de prazo atualizada.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao atualizar tarefa.");
    }
  }

  async function runRetentionPolicy() {
    if (user?.role !== "admin") return;
    setError("");
    setStatus("Executando política de retenção...");
    try {
      await apiFetch("/v1/compliance/retention/run", {
        method: "POST",
        headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
      });
      await Promise.all([fetchAILogs(), fetchPublications()]);
      setStatus("Política de retenção executada.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao executar retenção.");
    }
  }

  async function createUser(event) {
    event.preventDefault();
    if (user?.role !== "admin") return;
    setError("");
    if (newUserName.trim().length < 3) {
      setError("Nome do usuário deve ter ao menos 3 caracteres.");
      return;
    }
    try {
      await apiFetch("/v1/admin/users", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({
          name: newUserName.trim(),
          email: newUserEmail.trim(),
          role: newUserRole,
          status: "active"
        })
      });
      setNewUserName("");
      setNewUserEmail("");
      setNewUserRole("commercial");
      await fetchUsers();
      setStatus("Usuario cadastrado com sucesso.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha no cadastro de usuário.");
    }
  }

  async function updateUserStatus(userId, statusValue) {
    if (user?.role !== "admin") return;
    setError("");
    try {
      await apiFetch(`/v1/admin/users/${userId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({ status: statusValue })
      });
      await fetchUsers();
      setStatus("Status de usuário atualizado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao atualizar status de usuário.");
    }
  }

  async function updateUserRole(userId, roleValue) {
    if (user?.role !== "admin") return;
    setError("");
    try {
      await apiFetch(`/v1/admin/users/${userId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          ...(sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {})
        },
        body: JSON.stringify({ role: roleValue })
      });
      await fetchUsers();
      setStatus("Perfil de usuário atualizado.");
    } catch (err) {
      setError(err.message);
      setStatus("Falha ao atualizar perfil de usuário.");
    }
  }

  async function logout() {
    try {
      await apiFetch("/v1/identity/logout", {
        method: "POST",
        headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : undefined
      });
    } catch {}
    setSessionToken("");
    setUser(null);
    setOtpRequested(false);
    setOtp("");
    setLeads([]);
    setConversations([]);
    setConversationDrafts({});
    setTemplates([]);
    setFollowUps([]);
    setPublications([]);
    setPublicationValidationDrafts({});
    setDeadlineTasks([]);
    setDeadlineAlerts([]);
    setAiLogs([]);
    setUsers([]);
    setDevToken("");
    setShowLeadForm(false);
    setActivePage("dashboard");
    cancelEditLead();
    setStatus("Sessão encerrada.");
    setError("");
  }

  if (booting) {
    return (
      <main className="auth-shell">
        <section className="auth-card">
          <div className="auth-toolbar">
            <div className="brand">
              <BrandLogo dark={isDarkMode} />
            </div>
            <div className="auth-theme-switch">
              <ThemeSwitch themePreference={themePreference} onChange={handleThemeChange} />
            </div>
          </div>
          <p className="subtitle">Restaurando sessão...</p>
        </section>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="auth-shell">
        <section className="auth-card">
          <div className="auth-toolbar">
            <div className="brand">
              <BrandLogo dark={isDarkMode} />
            </div>
            <div className="auth-theme-switch">
              <ThemeSwitch themePreference={themePreference} onChange={handleThemeChange} />
            </div>
          </div>

          <h2>Acesse com e-mail cadastrado</h2>
          <p className="subtitle">
            Use o e-mail liberado para o piloto. Nao existe auto-cadastro publico por seguranca.
          </p>

          <form onSubmit={requestOTP} className="form" data-testid="login-form">
            <label>
              E-mail
              <input
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="admin@juriscan.local"
                required
              />
            </label>
            <button type="submit">Enviar código</button>
          </form>

          {otpRequested && (
            <form onSubmit={verifyOTP} className="form" data-testid="otp-form">
              <label>
                Código OTP
                <input
                  value={otp}
                  onChange={(event) => setOtp(onlyDigits(event.target.value).slice(0, 6))}
                  inputMode="numeric"
                  pattern="[0-9]{6}"
                  maxLength={6}
                  placeholder="123456"
                  required
                />
              </label>
              <button type="submit">Entrar</button>
            </form>
          )}

          {devToken && (
            <p className="hint" data-testid="dev-token-hint">
              Codigo de teste do piloto: <strong>{devToken}</strong>
              <button type="button" className="ghost-btn mini inline-action" onClick={() => setOtp(devToken)}>
                Usar codigo
              </button>
            </p>
          )}

          {!!status && <p className="status-line">Status: {status}</p>}
          {error && <p className="error">{error}</p>}
        </section>
      </main>
    );
  }

  const menuSections = [
    {
      label: "Comercial",
      items: [
        { key: "dashboard", label: "Entrada do Dia" },
        { key: "whatsapp", label: "WhatsApp" },
        { key: "leads-anchor", label: "Leads", target: "dashboard", anchor: "leads-pipeline" }
      ]
    },
    {
      label: "Operacional",
      items: ["admin", "controller", "lawyer"].includes(user.role)
        ? [
            { key: "publications", label: "Publicações" },
            { key: "deadlines", label: "Prazos" }
          ]
        : []
    },
    {
      label: "Controle",
      items: [
        ...(["admin", "controller"].includes(user.role) ? [{ key: "compliance", label: "Conferencia IA" }] : []),
        ...(user.role === "admin" ? [{ key: "users", label: "Usuarios" }] : [])
      ]
    }
  ].filter((section) => section.items.length > 0);
  const pageTitle = {
    dashboard: "Entrada do Dia",
    whatsapp: "WhatsApp",
    followups: "Retornos dos leads",
    publications: "Publicações",
    deadlines: "Prazos",
    compliance: "Conferencia IA",
    users: "Usuarios"
  }[activePage] || "Entrada do Dia";
  const pageSubtitle = {
    dashboard: "Filas simples para decidir o que responder, analisar e validar agora.",
    whatsapp: "Entrada de mensagens que podem virar lead com controle humano.",
    followups: "Retornos comerciais e modelos de mensagem dentro da rotina de leads.",
    publications: "Leitura assistida de publicações com validação humana.",
    deadlines: "Controle de prazos, responsáveis e alertas operacionais.",
    compliance: "Rastreabilidade das analises feitas por IA.",
    users: "Cadastro interno e permissoes de acesso."
  }[activePage] || "Filas simples para decidir o que responder, analisar e validar agora.";

  function navigateMenuItem(item) {
    const targetPage = item.target || item.key;
    setActivePage(targetPage);
    if (item.anchor) {
      window.setTimeout(() => {
        document.getElementById(item.anchor)?.scrollIntoView({ behavior: "smooth", block: "start" });
      }, 50);
    }
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <BrandLogo compact dark={isDarkMode} />
        </div>
        <nav className="side-nav">
          {menuSections.map((section) => (
            <div className="side-nav-section" key={section.label}>
              <span className="side-nav-label">{section.label}</span>
              {section.items.map((item) => (
                <button
                  key={item.key}
                  type="button"
                  data-testid={`nav-${item.key}`}
                  className={`side-nav-item ${activePage === item.key ? "active" : ""}`}
                  onClick={() => navigateMenuItem(item)}
                >
                  {item.label}
                </button>
              ))}
            </div>
          ))}
        </nav>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div className="brand-copy">
            <strong>{pageTitle}</strong>
            <p>{pageSubtitle}</p>
          </div>
          <div className="topbar-actions">
            <span className="user-chip">{user.email}</span>
            <ThemeSwitch themePreference={themePreference} onChange={handleThemeChange} compact />
            <button type="button" className="ghost-btn" onClick={logout}>
              Sair
            </button>
          </div>
        </header>

        {activePage === "dashboard" && (
          <div data-testid="page-dashboard">
          <>
            <section className="panel workbench-panel">
              <div className="panel-head">
                <div>
                  <h2>Mesa da controller</h2>
                  <p>Comece pelo que chegou hoje. A IA ajuda, mas prazo e classificacao sensivel precisam de validacao humana.</p>
                </div>
                <button type="button" className="ghost-btn" onClick={() => refreshDashboardData({ onlyOverdue: false })}>
                  Atualizar tudo
                </button>
              </div>
              <div className="queue-grid">
                <WorkQueueCard
                  label="1. WhatsApp"
                  title="Mensagens sem lead"
                  detail="Cole a conversa recebida e gere o lead quando fizer sentido."
                  count={pendingConversationsCount}
                  action="Abrir WhatsApp"
                  tone="info"
                  onClick={() => setActivePage("whatsapp")}
                />
                <WorkQueueCard
                  label="2. Comercial"
                  title="Leads aguardando resposta"
                  detail="Priorize SLA estourado, rode triagem IA e defina proximo passo."
                  count={openQueueItems.length}
                  action="Ver leads"
                  tone={overdueCount > 0 ? "danger" : "warn"}
                  onClick={() => setActivePage("followups")}
                />
                {["admin", "controller", "lawyer"].includes(user.role) && (
                  <WorkQueueCard
                    label="3. Publicacoes"
                    title="Textos para analisar"
                    detail="Registre publicacoes/despachos para sugestao de prazo."
                    count={pendingPublicationsCount}
                    action="Abrir publicacoes"
                    tone="info"
                    onClick={() => setActivePage("publications")}
                  />
                )}
                {["admin", "controller", "lawyer"].includes(user.role) && (
                  <WorkQueueCard
                    label="4. Validacao"
                    title="Prazos para conferir"
                    detail="Confirme data, responsavel e observacao antes de criar tarefa."
                    count={pendingValidationsCount}
                    action="Validar prazo"
                    tone="warn"
                    onClick={() => setActivePage("publications")}
                  />
                )}
                {["admin", "controller", "lawyer"].includes(user.role) && (
                  <WorkQueueCard
                    label="5. Operacao"
                    title="Tarefas de prazo abertas"
                    detail="Acompanhe cobrancas e riscos sem depender de planilha."
                    count={openDeadlineCount}
                    action="Ver prazos"
                    tone="success"
                    onClick={() => setActivePage("deadlines")}
                  />
                )}
                <WorkQueueCard
                  label="6. Retornos"
                  title="Retornos pendentes"
                  detail="Retornos comerciais programados para nao perder contato."
                  count={pendingFollowUpsCount}
                  action="Ver retornos"
                  tone="info"
                  onClick={() => setActivePage("followups")}
                />
              </div>
              <div className="security-strip">
                <span>Seguranca do piloto</span>
                <strong>Sem auto-cadastro publico.</strong>
                <strong>IA sugere; pessoa valida.</strong>
                <strong>Conferencia registra analises.</strong>
              </div>
            </section>

            <section className="stats-grid">
              <article className="stat-card">
                <h3>Total de leads</h3>
                <strong>{leadStats.total}</strong>
              </article>
              <article className="stat-card">
                <h3>Novos</h3>
                <strong>{leadStats.novos}</strong>
              </article>
              <article className="stat-card">
                <h3>Qualificados</h3>
                <strong>{leadStats.qualificados}</strong>
              </article>
              <article className="stat-card">
                <h3>Fechados</h3>
                <strong>{leadStats.fechados}</strong>
              </article>
            </section>

            <section className="panel" id="leads-pipeline">
              <div className="panel-head">
                <div>
                  <h2>Fila de resposta (SLA)</h2>
                  <p>Tempo máximo configurado: {slaMinutes} min para primeiro atendimento.</p>
                </div>
                <div className="panel-actions">
                  <label className="inline-check">
                    <input
                      type="checkbox"
                      checked={showOnlyOverdue}
                      onChange={async (event) => {
                        const checked = event.target.checked;
                        setShowOnlyOverdue(checked);
                        await fetchLeads({ onlyOverdue: checked });
                      }}
                    />
                    Mostrar apenas SLA estourado
                  </label>
                  <button
                    type="button"
                    className="ghost-btn"
                    onClick={() => fetchLeads({ onlyOverdue: showOnlyOverdue })}
                  >
                    Atualizar fila
                  </button>
                </div>
              </div>
              <div className="sla-summary">
                <span className="tag tag-warn">Em aberto: {openQueueItems.length}</span>
                <span className="tag tag-danger">Estourado: {overdueCount}</span>
              </div>
            </section>

            <section className="panel">
              <div className="panel-head">
                <div>
                  <h2>Leads</h2>
                  <p>Pipeline comercial do escritorio com controle completo dos contatos.</p>
                </div>
                <div className="panel-actions">
                  <button type="button" data-testid="open-create-lead" onClick={() => setShowLeadForm((v) => !v)}>
                    {showLeadForm ? "Fechar cadastro" : "Cadastrar lead"}
                  </button>
                  <button
                    type="button"
                    className="ghost-btn"
                    onClick={() => fetchLeads({ onlyOverdue: showOnlyOverdue })}
                  >
                    Atualizar
                  </button>
                </div>
              </div>

        {showLeadForm && (
          <form onSubmit={createLead} className="form boxed" data-testid="create-lead-form">
            <h3>Novo lead</h3>
            <label>
              Nome
              <input value={leadName} onChange={(event) => setLeadName(event.target.value)} minLength={3} required />
            </label>
            <label>
              Telefone
              <input
                value={leadPhone}
                onChange={(event) => setLeadPhone(formatPhoneBR(event.target.value))}
                inputMode="numeric"
                placeholder="(11) 99999-9999"
                maxLength={15}
                required
              />
            </label>
            <label>
              Origem
              <input value={leadOrigin} onChange={(event) => setLeadOrigin(event.target.value)} />
            </label>
            <label>
              Assunto
              <input value={leadSubject} onChange={(event) => setLeadSubject(event.target.value)} />
            </label>
            <label>
              Responsável
              <input
                type="email"
                value={leadOwnerEmail}
                onChange={(event) => setLeadOwnerEmail(event.target.value)}
                required
              />
            </label>
            <button type="submit">Salvar lead</button>
          </form>
        )}

        {editingLeadId && (
          <form onSubmit={updateLead} className="form boxed edit-box" data-testid="edit-lead-form">
            <h3>Editar lead</h3>
            <label>
              Nome
              <input value={editName} onChange={(event) => setEditName(event.target.value)} minLength={3} required />
            </label>
            <label>
              Telefone
              <input
                value={editPhone}
                onChange={(event) => setEditPhone(formatPhoneBR(event.target.value))}
                inputMode="numeric"
                placeholder="(11) 99999-9999"
                maxLength={15}
                required
              />
            </label>
            <label>
              Origem
              <input value={editOrigin} onChange={(event) => setEditOrigin(event.target.value)} />
            </label>
            <label>
              Assunto
              <input value={editSubject} onChange={(event) => setEditSubject(event.target.value)} />
            </label>
            <label>
              Responsável
              <input
                type="email"
                value={editOwnerEmail}
                onChange={(event) => setEditOwnerEmail(event.target.value)}
                required
              />
            </label>
            <label>
              Etapa
              <select value={editStage} onChange={(event) => setEditStage(event.target.value)}>
                {STAGES.map((stage) => (
                  <option key={stage} value={stage}>
                    {STAGE_LABEL[stage]}
                  </option>
                ))}
              </select>
            </label>
            <div className="panel-actions">
              <button type="submit">Salvar edição</button>
              <button type="button" className="ghost-btn" onClick={cancelEditLead}>
                Cancelar
              </button>
            </div>
          </form>
        )}

              <div className="pipeline-grid">
                {STAGES.map((stage) => (
                  <section key={stage} className="pipeline-column" data-testid={`column-${stage}`}>
                    <h3>{STAGE_LABEL[stage]}</h3>
                    {leadsByStage[stage].length === 0 ? (
                      <p className="muted">Sem leads.</p>
                    ) : (
                      <ul className="lead-list">
                        {leadsByStage[stage].map((lead) => (
                          <li key={lead.id} className="lead-item" data-testid="lead-item">
                            <div className="lead-head">
                              <strong>{lead.name}</strong>
                              <span className={statusTagClass(resolveLeadSLAStatus(lead))}>
                                {resolveLeadSLAStatus(lead)}
                              </span>
                            </div>
                            <span>{formatPhoneBR(lead.phone)}</span>
                            <span>{lead.subject || "Sem assunto"}</span>
                            <span className="muted small-history">
                              Sem resposta: {resolveLeadMinutesWithoutResponse(lead)} min
                            </span>
                            <span className="muted small-history">
                              Histórico: {Array.isArray(lead.history) ? lead.history.length : 0} evento(s)
                            </span>
                            {lead.ai_classification && (
                              <div className="ai-chip-group">
                                <span className="tag tag-info">
                                  IA: {lead.ai_classification.category} / {lead.ai_classification.urgency}
                                </span>
                                <span className="tag tag-muted">Score: {lead.ai_classification.score}</span>
                              </div>
                            )}
                            {lead.next_step && <span className="muted small-history">Próximo passo: {lead.next_step}</span>}
                            {lead.next_follow_up_at && (
                              <span className="muted small-history">
                                Retorno: {new Date(lead.next_follow_up_at).toLocaleString()}
                              </span>
                            )}
                            <label>
                              Mover para
                              <select value={lead.stage} onChange={(event) => changeLeadStage(lead, event.target.value)}>
                                {STAGES.map((value) => (
                                  <option key={value} value={value}>
                                    {STAGE_LABEL[value]}
                                  </option>
                                ))}
                              </select>
                            </label>
                            <div className="lead-actions">
                              <button type="button" className="ghost-btn mini" onClick={() => runLeadTriage(lead.id)}>
                                Triar IA
                              </button>
                              {lead.ai_classification?.suggested_action && (
                                <button
                                  type="button"
                                  className="ghost-btn mini"
                                  onClick={() => applySuggestedNextStep(lead)}
                                >
                                  Aplicar sugestão
                                </button>
                              )}
                              {lead.ai_classification && ["admin", "controller", "lawyer"].includes(user.role) && (
                                <button
                                  type="button"
                                  className="ghost-btn mini"
                                  onClick={() => overrideLeadTriage(lead)}
                                >
                                  Override IA
                                </button>
                              )}
                              <button type="button" className="ghost-btn mini" onClick={() => startEditLead(lead)}>
                                Editar
                              </button>
                            </div>
                          </li>
                        ))}
                      </ul>
                    )}
                  </section>
                ))}
              </div>
            </section>
          </>
          </div>
        )}

        {activePage === "whatsapp" && (
          <section className="panel" data-testid="page-whatsapp">
        <div className="panel-head">
          <div>
            <h2>WhatsApp</h2>
            <p>Entrada comercial para registrar mensagem recebida e transformar em lead quando fizer sentido.</p>
          </div>
          <button type="button" className="ghost-btn" onClick={fetchConversations}>
            Atualizar caixa
          </button>
        </div>

        <form className="form boxed" onSubmit={simulateWhatsAppMessage}>
          <h3>Registrar mensagem recebida</h3>
          <label>
            Telefone
            <input
              value={simPhone}
              onChange={(event) => setSimPhone(formatPhoneBR(event.target.value))}
              inputMode="numeric"
              placeholder="(11) 99999-9999"
              maxLength={15}
              required
            />
          </label>
          <label>
            Nome do contato
            <input value={simName} onChange={(event) => setSimName(event.target.value)} placeholder="Opcional" />
          </label>
          <label>
            Mensagem
            <input value={simMessage} onChange={(event) => setSimMessage(event.target.value)} required />
          </label>
          <label className="inline-check">
            <input
              type="checkbox"
              checked={simCreateLead}
              onChange={(event) => setSimCreateLead(event.target.checked)}
            />
            Gerar lead automaticamente
          </label>
          <button type="submit">{simCreateLead ? "Receber e gerar lead" : "Receber na caixa"}</button>
        </form>

        <div className="conversation-list">
          {conversations.length === 0 && <p className="muted">Sem conversas no momento.</p>}
          {conversations.map((conversation) => {
            const draft = conversationDrafts[conversation.id] || {
              status: conversation.status,
              leadID: conversation.lead_id || ""
            };
            return (
              <article className="conversation-item" key={conversation.id}>
                <div className="conversation-head">
                  <div>
                    <strong>{conversation.contact_name || formatPhoneBR(conversation.phone)}</strong>
                    <p className="muted">{formatPhoneBR(conversation.phone)}</p>
                  </div>
                  <span className="tag tag-info">{CONVERSATION_STATUS_LABEL[conversation.status] || conversation.status}</span>
                </div>
                <p>{conversation.last_message}</p>
                <p className="muted">Mensagens: {conversation.message_count}</p>
                {conversation.suggested_lead_id && (
                  <p className="muted">
                    Sugestão automática: {conversation.suggested_lead_name || conversation.suggested_lead_id}
                  </p>
                )}
                <div className="conversation-actions">
                  <label>
                    Classificação
                    <select
                      value={draft.status}
                      onChange={(event) => updateConversationDraft(conversation.id, "status", event.target.value)}
                    >
                      <option value="nova">Nova</option>
                      <option value="sem_lead">Sem lead</option>
                      <option value="vinculada">Vinculada</option>
                    </select>
                  </label>
                  {draft.status === "vinculada" && (
                    <label>
                      Lead
                      <select
                        value={draft.leadID}
                        onChange={(event) => updateConversationDraft(conversation.id, "leadID", event.target.value)}
                      >
                        <option value="">Selecione</option>
                        {leads.map((lead) => (
                          <option key={lead.id} value={lead.id}>
                            {lead.name}
                          </option>
                        ))}
                      </select>
                    </label>
                  )}
                  <button type="button" className="ghost-btn" onClick={() => saveConversationClassification(conversation.id)}>
                    Salvar
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      </section>
        )}

        {activePage === "followups" && (
          <section className="panel" data-testid="page-followups">
            <div className="panel-head">
              <div>
                <h2>Retornos dos leads</h2>
                <p>Agenda de retorno e modelos de mensagem. O pipeline principal fica em Leads.</p>
              </div>
              <button type="button" className="ghost-btn" onClick={() => Promise.all([fetchTemplates(), fetchFollowUps()])}>
                Atualizar
              </button>
            </div>

            {["admin", "commercial"].includes(user.role) && (
              <form className="form boxed" onSubmit={createTemplate}>
                <h3>Novo modelo de resposta</h3>
                <label>
                  Nome
                  <input
                    value={newTemplateName}
                    onChange={(event) => setNewTemplateName(event.target.value)}
                    placeholder="Ex: Primeiro contato"
                    required
                  />
                </label>
                <label>
                  Mensagem
                  <textarea
                    value={newTemplateBody}
                    onChange={(event) => setNewTemplateBody(event.target.value)}
                    placeholder="Mensagem padrao para WhatsApp"
                    rows={3}
                    required
                  />
                </label>
                <button type="submit">Salvar template</button>
              </form>
            )}

            <form className="form boxed" onSubmit={scheduleFollowUp}>
              <h3>Agendar retorno</h3>
              <label>
                Lead
                <select value={followUpLeadID} onChange={(event) => setFollowUpLeadID(event.target.value)} required>
                  <option value="">Selecione</option>
                  {leads.map((lead) => (
                    <option key={lead.id} value={lead.id}>
                      {lead.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Template (opcional)
                <select
                  value={followUpTemplateID}
                  onChange={(event) => {
                    const templateID = event.target.value;
                    setFollowUpTemplateID(templateID);
                    const tpl = templates.find((item) => item.id === templateID);
                    if (tpl) setFollowUpMessage(tpl.body);
                  }}
                >
                  <option value="">Selecione</option>
                  {templates.map((template) => (
                    <option key={template.id} value={template.id}>
                      {template.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Mensagem
                <textarea
                  value={followUpMessage}
                  onChange={(event) => setFollowUpMessage(event.target.value)}
                  rows={3}
                  required
                />
              </label>
              <label>
                Data e hora do retorno
                <input
                  type="datetime-local"
                  value={followUpDueAt}
                  onChange={(event) => setFollowUpDueAt(event.target.value)}
                  required
                />
              </label>
              <button type="submit">Agendar retorno</button>
            </form>

            <div className="dual-grid">
              <article className="panel-sub">
                <h3>Templates cadastrados</h3>
                {templates.length === 0 && <p className="muted">Sem templates cadastrados.</p>}
                <ul className="simple-list">
                  {templates.map((template) => (
                    <li key={template.id}>
                      <strong>{template.name}</strong>
                      <p className="muted">{template.body}</p>
                    </li>
                  ))}
                </ul>
              </article>
              <article className="panel-sub">
                <h3>Retornos agendados</h3>
                {followUps.length === 0 && <p className="muted">Sem retornos agendados.</p>}
                <ul className="simple-list">
                  {followUps.map((item) => (
                    <li key={item.id}>
                      <div className="list-head">
                        <strong>{leads.find((lead) => lead.id === item.lead_id)?.name || item.lead_id}</strong>
                        <span className={`tag ${item.status === "pendente" ? "tag-warn" : "tag-success"}`}>
                          {FOLLOW_UP_STATUS_LABEL[item.status] || item.status}
                        </span>
                      </div>
                      <p>{item.message}</p>
                      <p className="muted">Prazo: {new Date(item.due_at).toLocaleString()}</p>
                      <div className="panel-actions">
                        {item.status === "pendente" && (
                          <>
                            <button
                              type="button"
                              className="ghost-btn mini"
                              onClick={() => markFollowUpStatus(item.id, "concluido")}
                            >
                              Concluir
                            </button>
                            <button
                              type="button"
                              className="ghost-btn mini"
                              onClick={() => markFollowUpStatus(item.id, "cancelado")}
                            >
                              Cancelar
                            </button>
                          </>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              </article>
            </div>
          </section>
        )}

        {activePage === "publications" && (
          <section className="panel" data-testid="page-publications">
            <div className="panel-head">
              <div>
                <h2>Publicações</h2>
                <p>Ingestão de texto/arquivo com análise IA e validação humana obrigatória.</p>
              </div>
              <button type="button" className="ghost-btn" onClick={fetchPublications}>
                Atualizar publicações
              </button>
            </div>

            <form className="form boxed" onSubmit={createPublication}>
              <h3>Registrar publicação</h3>
              <label>
                Origem
                <input value={publicationSource} onChange={(event) => setPublicationSource(event.target.value)} required />
              </label>
              <label>
                Tipo de entrada
                <select
                  value={publicationInputType}
                  onChange={(event) => setPublicationInputType(event.target.value)}
                >
                  <option value="texto">Texto</option>
                  <option value="arquivo">Arquivo</option>
                </select>
              </label>
              {publicationInputType === "arquivo" && (
                <label>
                  Nome do arquivo
                  <input
                    value={publicationFileName}
                    onChange={(event) => setPublicationFileName(event.target.value)}
                    placeholder="publicacao-2026-04-14.pdf"
                  />
                </label>
              )}
              <label>
                Conteúdo
                <textarea
                  value={publicationRawText}
                  onChange={(event) => setPublicationRawText(event.target.value)}
                  rows={5}
                  required
                />
              </label>
              <button type="submit">Salvar publicação</button>
            </form>

            <div className="conversation-list">
              {publications.length === 0 && <p className="muted">Sem publicações registradas.</p>}
              {publications.map((item) => {
                const draft = publicationValidationDrafts[item.id] || {
                  finalDeadlineAt: toDateTimeLocalInput(item.analysis?.suggested_deadline_at),
                  notes: "",
                  ownerEmail: item.analysis?.suggested_owner_email || ""
                };
                return (
                  <article key={item.id} className="conversation-item">
                    <div className="conversation-head">
                      <div>
                        <strong>{item.source}</strong>
                        <p className="muted">{item.input_type === "arquivo" ? item.file_name || "Arquivo" : "Texto"}</p>
                      </div>
                      <span className="tag tag-info">{item.analysis ? "Analisado" : "Pendente análise"}</span>
                    </div>
                    <p className="muted">{item.raw_text}</p>
                    <div className="panel-actions">
                      <button type="button" className="ghost-btn mini" onClick={() => analyzePublication(item.id)}>
                        Analisar IA
                      </button>
                    </div>

                    {item.analysis && (
                      <div className="analysis-box">
                        <p>
                          <strong>Ato:</strong> {item.analysis.act_type} | <strong>Risco:</strong> {item.analysis.risk}
                        </p>
                        <p>
                          <strong>Prazo sugerido:</strong>{" "}
                          {new Date(item.analysis.suggested_deadline_at).toLocaleString()} ({item.analysis.suggested_deadline_days} dias)
                        </p>
                        <p className="muted">Confiança IA: {Math.round((item.analysis.confidence || 0) * 100)}%</p>

                        {!item.validation && (
                          <div className="validation-grid">
                            <label>
                              Prazo final (humano)
                              <input
                                type="datetime-local"
                                value={draft.finalDeadlineAt}
                                onChange={(event) =>
                                  updatePublicationValidationDraft(item.id, "finalDeadlineAt", event.target.value)
                                }
                              />
                            </label>
                            <label>
                              Dono da tarefa
                              <input
                                type="email"
                                value={draft.ownerEmail}
                                onChange={(event) =>
                                  updatePublicationValidationDraft(item.id, "ownerEmail", event.target.value)
                                }
                              />
                            </label>
                            <label>
                              Observações
                              <textarea
                                value={draft.notes}
                                rows={2}
                                onChange={(event) =>
                                  updatePublicationValidationDraft(item.id, "notes", event.target.value)
                                }
                              />
                            </label>
                            <button type="button" onClick={() => validatePublication(item.id)}>
                              Confirmar prazo e criar tarefa
                            </button>
                          </div>
                        )}

                        {item.validation && (
                          <p className="muted">
                            Prazo validado por {item.validation.validated_by} em{" "}
                            {new Date(item.validation.validated_at).toLocaleString()}
                          </p>
                        )}
                      </div>
                    )}
                  </article>
                );
              })}
            </div>
          </section>
        )}

        {activePage === "deadlines" && (
          <section className="panel" data-testid="page-deadlines">
            <div className="panel-head">
              <div>
                <h2>Prazos</h2>
                <p>Operação diária com dono, status e alertas de vencimento.</p>
              </div>
              <button
                type="button"
                className="ghost-btn"
                onClick={() => Promise.all([fetchDeadlineTasks(), fetchDeadlineAlerts()])}
              >
                Atualizar
              </button>
            </div>

            <div className="sla-summary">
              {deadlineAlerts.length === 0 ? (
                <span className="tag tag-success">Sem alertas ativos</span>
              ) : (
                deadlineAlerts.map((alert) => (
                  <span key={alert.id} className={`tag ${alert.type === "atrasado" ? "tag-danger" : "tag-warn"}`}>
                    {alert.message}
                  </span>
                ))
              )}
            </div>

            <div className="conversation-list">
              {deadlineTasks.length === 0 && <p className="muted">Sem tarefas de prazo.</p>}
              {deadlineTasks.map((task) => (
                <article key={task.id} className="conversation-item">
                  <div className="conversation-head">
                    <div>
                      <strong>{task.title}</strong>
                      <p className="muted">
                        Vence em {new Date(task.due_at).toLocaleString()} | Responsável: {task.owner_email}
                      </p>
                    </div>
                    <span className="tag tag-info">{task.risk}</span>
                  </div>
                  <div className="conversation-actions">
                    <label>
                      Status
                      <select
                        value={task.status}
                        onChange={(event) => updateDeadlineTask(task.id, { status: event.target.value })}
                      >
                        {DEADLINE_STATUS_OPTIONS.map((option) => (
                          <option key={option.value} value={option.value}>
                            {option.label}
                          </option>
                        ))}
                      </select>
                    </label>
                  </div>
                </article>
              ))}
            </div>
          </section>
        )}

        {activePage === "compliance" && ["admin", "controller"].includes(user.role) && (
          <section className="panel" data-testid="page-compliance">
            <div className="panel-head">
              <div>
                <h2>Conferencia IA</h2>
                <p>Rastreio de prompt/resposta, modelo e confianca por analise de IA.</p>
              </div>
              <div className="panel-actions">
                <button type="button" className="ghost-btn" onClick={fetchAILogs}>
                  Atualizar logs
                </button>
                {user.role === "admin" && (
                  <button type="button" className="ghost-btn" onClick={runRetentionPolicy}>
                    Executar retenção
                  </button>
                )}
              </div>
            </div>

            <div className="table-wrap">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Quando</th>
                    <th>Feature</th>
                    <th>Recurso</th>
                    <th>Modelo</th>
                    <th>Confiança</th>
                    <th>Retenção até</th>
                  </tr>
                </thead>
                <tbody>
                  {aiLogs.length === 0 && (
                    <tr>
                      <td colSpan={6} className="muted">
                        Sem logs de IA.
                      </td>
                    </tr>
                  )}
                  {aiLogs.map((item) => (
                    <tr key={item.id}>
                      <td>{new Date(item.created_at).toLocaleString()}</td>
                      <td>{item.feature}</td>
                      <td>
                        {item.resource_type} / {item.resource_id}
                      </td>
                      <td>{item.model}</td>
                      <td>{Math.round((item.confidence || 0) * 100)}%</td>
                      <td>{new Date(item.retention_until).toLocaleDateString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {activePage === "users" && user.role === "admin" && (
          <section className="panel" data-testid="page-users">
          <div className="panel-head">
            <div>
              <h2>Usuarios</h2>
              <p>Cadastro interno sem auto-cadastro publico. Apenas usuario ativo pode logar.</p>
            </div>
            <button type="button" className="ghost-btn" onClick={fetchUsers}>
              Atualizar usuários
            </button>
          </div>

          <form className="form boxed" onSubmit={createUser}>
            <h3>Novo usuário</h3>
            <label>
              Nome
              <input value={newUserName} onChange={(event) => setNewUserName(event.target.value)} required />
            </label>
            <label>
              E-mail
              <input
                type="email"
                value={newUserEmail}
                onChange={(event) => setNewUserEmail(event.target.value)}
                required
              />
            </label>
            <label>
              Perfil
              <select value={newUserRole} onChange={(event) => setNewUserRole(event.target.value)}>
                {USER_ROLES.map((role) => (
                  <option key={role.value} value={role.value}>
                    {role.label}
                  </option>
                ))}
              </select>
            </label>
            <button type="submit">Cadastrar usuário</button>
          </form>

          <div className="conversation-list">
            {users.length === 0 && <p className="muted">Sem usuários cadastrados.</p>}
            {users.map((item) => (
              <article key={item.id} className="conversation-item">
                <div className="conversation-head">
                  <div>
                    <strong>{item.name}</strong>
                    <p className="muted">{item.email}</p>
                  </div>
                  <div className="user-tags">
                    <span className="tag tag-info">{resolveUserRoleLabel(item.role)}</span>
                    <span className={`tag ${item.status === "active" ? "tag-success" : "tag-muted"}`}>
                      {resolveUserStatusLabel(item.status)}
                    </span>
                  </div>
                </div>
                <p className="muted">
                  Último login: {item.last_login_at ? new Date(item.last_login_at).toLocaleString() : "nunca"}
                </p>
                <div className="conversation-actions">
                  <label>
                    Perfil
                    <select
                      value={item.role}
                      onChange={(event) => updateUserRole(item.id, event.target.value)}
                    >
                      {USER_ROLES.map((role) => (
                        <option key={role.value} value={role.value}>
                          {role.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  {item.status === "active" ? (
                    <button type="button" className="ghost-btn" onClick={() => updateUserStatus(item.id, "inactive")}>
                      Inativar
                    </button>
                  ) : (
                    <button type="button" className="ghost-btn" onClick={() => updateUserStatus(item.id, "active")}>
                      Reativar
                    </button>
                  )}
                </div>
              </article>
            ))}
          </div>
        </section>
        )}

        {!!status && <p className="status-line">Status: {status}</p>}
        {error && <p className="error">{error}</p>}
      </section>
    </main>
  );
}
