# AGENTS.md instructions for this repo

## CLI operations

Before running commands, generating automation, or interpreting reports, read
[Platform API agent guidance](docs/PLATFORM_API_AGENT_GUIDE.md). Use the docs and
binary from the same checkout/version; an installed CLI may predate this migration.

## Skills
A skill is a set of local instructions stored in a `SKILL.md` file.

### Available skills
- apple-ads-certification-core: Core Apple Ads certification knowledge for campaign structure, keyword strategy, and performance interpretation. Use for planning/review requests that ask for certification-aligned guidance. (file: skills/apple-ads-certification-core/SKILL.md)
- apple-ads-discovery-promotion: Discovery -> Promotion -> Isolation workflow skill for turning search terms into exact-match keywords with negative-keyword isolation. Use for weekly optimization loops and automation specs. (file: skills/apple-ads-discovery-promotion/SKILL.md)
- apple-ads-placements-creative: Apple Ads placement and creative certification guidance (Today tab, Search tab, Search results, Product pages, custom product pages, variation behavior). Use for ad placement/creative setup decisions. (file: skills/apple-ads-placements-creative/SKILL.md)

### How to use skills
- Discovery: The list above is the skills available in this repo.
- Trigger rules: If the user names a skill (with `$SkillName` or plain text) or the task clearly matches a skill description, use that skill.
- Progressive disclosure:
  1. Open the skill `SKILL.md` first.
  2. Load only the specific `references/` files needed for the current task.
  3. Avoid bulk-loading all reference files.
- Keep context lean: summarize long references; only pull detailed sections needed to complete the task.
