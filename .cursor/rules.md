# NUNCA recrie arquivos sem confirmar explicitamente com você primeiro
- Se eu não conseguir ler um arquivo, devo usar cat, head, ou outros comandos de terminal
- Perguntar a você sobre o estado real dos arquivos
# Faça APENAS modificações incrementais
- Modificar o arquivo existente com search_replace
- Não reescrever tudo do zero
# Quando comandos não tem output:
- Adicionar && echo "OK" ou || echo "ERRO"
- Usar redirecionamento para arquivo e depois ler
- Testar com comandos simples primeiro
# Verificar estado real antes de agir:
- ls, cat, file para confirmar existência
- git status para ver mudanças
- Não assumir nada

