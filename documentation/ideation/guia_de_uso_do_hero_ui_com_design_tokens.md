# Guia de Uso do HeroUI com Design Tokens

Este documento define **como usar corretamente o HeroUI em conjunto com design tokens**, garantindo reutilização de UI entre **web-app (Next.js)** e **obsidian-plugin**, mantendo flexibilidade, consistência visual e independência de runtime.

---

## 1. Objetivo

- Usar **HeroUI como base de componentes React**
- Centralizar decisões visuais em **design tokens**
- Permitir múltiplos runtimes (web / Obsidian)
- Evitar acoplamento direto a cores, spacing ou temas fixos

> HeroUI é a camada de componentes.  
> Tokens são a camada de decisão visual.

---

## 2. O que são Design Tokens (na prática)

Design tokens são **valores semânticos**, não visuais diretos.

❌ Errado (valor visual direto)
```tsx
<Card className="bg-zinc-900 text-white" />
```

✅ Correto (valor semântico)
```tsx
<Card className="bg-[var(--se-surface)] text-[var(--se-text)]" />
```

Tokens permitem que o **mesmo componente** funcione em contextos diferentes.

---

## 3. Estrutura Recomendada do Projeto

```txt
apps/
 ├─ web-app/            # Next.js
 └─ obsidian-plugin/    # Plugin Obsidian

packages/
 ├─ ui-package/         # Componentes HeroUI + wrappers
 ├─ tokens/             # Design tokens (CSS + TS)
 └─ shared-ts/           # Tipos, DTOs, hooks
```

---

## 4. Tokens como CSS Variables (fonte da verdade)

### 4.1 Tokens base

```css
:root {
  --se-bg: #0b0b0f;
  --se-surface: #14141a;
  --se-text: #fafafa;
  --se-muted: #a1a1aa;

  --se-radius-sm: 6px;
  --se-radius-md: 10px;

  --se-space-xs: 4px;
  --se-space-sm: 8px;
  --se-space-md: 12px;
  --se-space-lg: 16px;
}
```

Esses tokens **não sabem onde estão sendo usados**.

---

### 4.2 Tokens para Obsidian

```css
.se-obsidian {
  --se-bg: var(--background-primary);
  --se-surface: var(--background-secondary);
  --se-text: var(--text-normal);
  --se-muted: var(--text-muted);
}
```

O plugin deve montar a UI dentro de um container:
```html
<div class="se-root se-obsidian"></div>
```

---

### 4.3 Tokens para Web App

```css
.se-web {
  --se-bg: #0b0b0f;
  --se-surface: #14141a;
  --se-text: #fafafa;
  --se-muted: #9ca3af;
}
```

---

## 5. HeroUI + Tokens (uso correto)

HeroUI aceita classes Tailwind customizadas.

### Exemplo correto
```tsx
import { Card } from "@heroui/react";

export function StoryCard({ children }) {
  return (
    <Card className="bg-[var(--se-surface)] text-[var(--se-text)] rounded-[var(--se-radius-md)] p-[var(--se-space-md)]">
      {children}
    </Card>
  );
}
```

### Regras
- ❌ Nunca usar cores fixas
- ❌ Nunca usar spacing hardcoded
- ✅ Sempre usar tokens

---

## 6. Wrappers no ui-package (padrão recomendado)

Nunca exporte HeroUI cru direto para os apps.

```tsx
// ui-package/components/SECard.tsx
import { Card } from "@heroui/react";

export function SECard(props) {
  return (
    <Card
      {...props}
      className={`
        bg-[var(--se-surface)]
        text-[var(--se-text)]
        rounded-[var(--se-radius-md)]
        ${props.className ?? ""}
      `}
    />
  );
}
```

Benefícios:
- Centraliza decisões visuais
- Evita espalhar tokens
- Facilita refactor

---

## 7. O que NÃO fazer

❌ Usar HeroUI com tema próprio sem tokens
❌ Misturar `bg-zinc-*` com CSS variables
❌ Criar tokens específicos por componente
❌ Copiar estilos do site do HeroUI

---

## 8. Tokens em TypeScript (opcional)

Para lógica (ex: spacing em JS):

```ts
export const spacing = {
  xs: "var(--se-space-xs)",
  sm: "var(--se-space-sm)",
  md: "var(--se-space-md)",
  lg: "var(--se-space-lg)",
};
```

Nunca use tokens TS para estilização visual direta — apenas apoio.

---

## 9. Checklist de Uso Correto

- [ ] Componentes não usam cores diretas
- [ ] UI funciona em dark/light
- [ ] Plugin respeita tema do Obsidian
- [ ] Web-app mantém identidade própria
- [ ] Tokens versionados

---

## 10. Conclusão

HeroUI resolve **componentes e interações**.
Tokens resolvem **consistência, escala e portabilidade**.

Separar essas responsabilidades é o que permite:
- Apps irmãs
- Evolução sem reescrita
- UI previsível no longo prazo

---

> Regra de ouro:  
> **Componentes não sabem cores. Tokens sabem.**

