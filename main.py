"""
Auto-Ban Bot — Banni automatiquement tout membre qui reçoit un rôle interdit.
Toute la configuration se gère via des commandes Discord (prefix !)
et est persistée dans config.json.
"""

import discord
from discord.ext import commands
import json
import os
import logging
from datetime import datetime, timezone

# ─── Logging console + fichier ───────────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("bot.log", encoding="utf-8"),
        logging.StreamHandler()
    ]
)

TOKEN = os.environ.get("DISCORD_TOKEN", "VOTRE_TOKEN_ICI")
CONFIG_FILE = "config.json"

# ─── Config par défaut (écrasée par config.json si présent) ──────────────────
DEFAULT_CONFIG = {
    "banned_roles": [],          # noms ou IDs de rôles qui déclenchent le ban
    "log_channel_id": None,      # ID du salon de log (int ou null)
    "ban_reason": "Ban automatique : rôle interdit détecté.",
    "delete_messages_days": 1,   # jours de messages supprimés lors du ban (0-7)
    "dm_user_before_ban": True,  # DM l'utilisateur avant de le bannir
    "dm_message": "Tu as été banni automatiquement du serveur pour avoir reçu un rôle non autorisé.",
    "ban_history": []            # liste des bans effectués par le bot
}


# ─── Persistance ─────────────────────────────────────────────────────────────

def load_config() -> dict:
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
        # Fusionner avec les défauts pour les clés manquantes
        for key, value in DEFAULT_CONFIG.items():
            data.setdefault(key, value)
        return data
    return dict(DEFAULT_CONFIG)


def save_config(cfg: dict):
    with open(CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(cfg, f, indent=2, ensure_ascii=False)


config = load_config()

# ─── Bot ─────────────────────────────────────────────────────────────────────

intents = discord.Intents.default()
intents.members = True
intents.message_content = True

bot = commands.Bot(command_prefix="!", intents=intents)


# ─── Helpers ─────────────────────────────────────────────────────────────────

def is_banned_role(role: discord.Role) -> bool:
    for entry in config["banned_roles"]:
        if isinstance(entry, int) and role.id == entry:
            return True
        if isinstance(entry, str) and role.name.lower() == entry.lower():
            return True
    return False


async def send_log(guild: discord.Guild, embed: discord.Embed):
    if config["log_channel_id"]:
        ch = guild.get_channel(int(config["log_channel_id"]))
        if ch:
            try:
                await ch.send(embed=embed)
            except discord.Forbidden:
                logging.warning("Impossible d'écrire dans le salon de log.")


async def ban_member(member: discord.Member, trigger_role_name: str):
    reason = f"{config['ban_reason']} (rôle : {trigger_role_name})"

    # DM avant le ban si activé
    if config["dm_user_before_ban"]:
        try:
            await member.send(config["dm_message"])
        except discord.Forbidden:
            logging.info(f"DM impossible pour {member} (DMs fermés).")

    try:
        await member.ban(
            reason=reason,
            delete_message_days=max(0, min(7, int(config["delete_messages_days"])))
        )
        timestamp = datetime.now(timezone.utc).isoformat()
        logging.info(f"Banni : {member} ({member.id}) — {reason}")

        # Historique
        config["ban_history"].append({
            "user": str(member),
            "user_id": member.id,
            "role": trigger_role_name,
            "reason": reason,
            "timestamp": timestamp
        })
        # Garder seulement les 200 derniers
        config["ban_history"] = config["ban_history"][-200:]
        save_config(config)

        # Embed de log
        embed = discord.Embed(
            title="🔨 Ban automatique",
            color=discord.Color.red(),
            timestamp=datetime.now(timezone.utc)
        )
        embed.add_field(name="Utilisateur", value=f"{member} (`{member.id}`)", inline=True)
        embed.add_field(name="Rôle déclencheur", value=trigger_role_name, inline=True)
        embed.add_field(name="Raison", value=reason, inline=False)
        embed.set_footer(text=f"DM envoyé : {'oui' if config['dm_user_before_ban'] else 'non'}")
        await send_log(member.guild, embed)

    except discord.Forbidden:
        logging.warning(f"Permission insuffisante pour bannir {member}.")
    except discord.HTTPException as e:
        logging.error(f"Erreur HTTP lors du ban de {member}: {e}")


# ─── Événements ──────────────────────────────────────────────────────────────

@bot.event
async def on_ready():
    logging.info(f"✅ Connecté en tant que {bot.user} ({bot.user.id})")
    logging.info(f"Rôles surveillés : {config['banned_roles']}")
    logging.info(f"Salon de log    : {config['log_channel_id']}")
    logging.info(f"DM avant ban    : {config['dm_user_before_ban']}")


@bot.event
async def on_member_update(before: discord.Member, after: discord.Member):
    added_roles = set(after.roles) - set(before.roles)
    for role in added_roles:
        if is_banned_role(role):
            await ban_member(after, role.name)
            break


@bot.event
async def on_member_join(member: discord.Member):
    for role in member.roles:
        if is_banned_role(role):
            await ban_member(member, role.name)
            break


# ─── Guard : commandes réservées aux admins ───────────────────────────────────

def admin_only():
    return commands.has_permissions(administrator=True)


# ════════════════════════════════════════════════════════════════════════════
#  COMMANDES — RÔLES INTERDITS
# ════════════════════════════════════════════════════════════════════════════

@bot.command(name="addrole")
@admin_only()
async def cmd_addrole(ctx, *, role_name: str):
    """!addrole <nom_du_rôle>  —  Ajoute un rôle à la liste des rôles interdits."""
    if role_name in config["banned_roles"]:
        await ctx.send(f"⚠️ `{role_name}` est déjà dans la liste.")
        return
    config["banned_roles"].append(role_name)
    save_config(config)
    await ctx.send(f"✅ Rôle **{role_name}** ajouté aux rôles interdits.")


@bot.command(name="removerole")
@admin_only()
async def cmd_removerole(ctx, *, role_name: str):
    """!removerole <nom_du_rôle>  —  Retire un rôle de la liste."""
    if role_name not in config["banned_roles"]:
        await ctx.send(f"❌ `{role_name}` introuvable dans la liste.")
        return
    config["banned_roles"].remove(role_name)
    save_config(config)
    await ctx.send(f"✅ Rôle **{role_name}** retiré des rôles interdits.")


@bot.command(name="listroles")
@admin_only()
async def cmd_listroles(ctx):
    """!listroles  —  Affiche les rôles actuellement surveillés."""
    if not config["banned_roles"]:
        await ctx.send("Aucun rôle interdit configuré.")
        return
    lines = "\n".join(f"• `{r}`" for r in config["banned_roles"])
    await ctx.send(f"**Rôles interdits ({len(config['banned_roles'])}) :**\n{lines}")


# ════════════════════════════════════════════════════════════════════════════
#  COMMANDES — SALON DE LOG
# ════════════════════════════════════════════════════════════════════════════

@bot.command(name="setlog")
@admin_only()
async def cmd_setlog(ctx, channel: discord.TextChannel = None):
    """!setlog [#salon]  —  Définit le salon de log. Sans argument = salon actuel."""
    target = channel or ctx.channel
    config["log_channel_id"] = target.id
    save_config(config)
    await ctx.send(f"✅ Salon de log défini sur {target.mention}.")


@bot.command(name="removelog")
@admin_only()
async def cmd_removelog(ctx):
    """!removelog  —  Désactive le salon de log."""
    config["log_channel_id"] = None
    save_config(config)
    await ctx.send("✅ Salon de log désactivé.")


# ════════════════════════════════════════════════════════════════════════════
#  COMMANDES — MESSAGE DE BAN / DM
# ════════════════════════════════════════════════════════════════════════════

@bot.command(name="setreason")
@admin_only()
async def cmd_setreason(ctx, *, reason: str):
    """!setreason <texte>  —  Modifie la raison affichée dans l'audit log Discord."""
    config["ban_reason"] = reason
    save_config(config)
    await ctx.send(f"✅ Raison de ban mise à jour :\n> {reason}")


@bot.command(name="setdm")
@admin_only()
async def cmd_setdm(ctx, *, message: str):
    """!setdm <message>  —  Modifie le DM envoyé à l'utilisateur avant le ban."""
    config["dm_message"] = message
    save_config(config)
    await ctx.send(f"✅ Message DM mis à jour :\n> {message}")


@bot.command(name="toggledm")
@admin_only()
async def cmd_toggledm(ctx):
    """!toggledm  —  Active/désactive le DM envoyé avant le ban."""
    config["dm_user_before_ban"] = not config["dm_user_before_ban"]
    save_config(config)
    state = "activé ✅" if config["dm_user_before_ban"] else "désactivé ❌"
    await ctx.send(f"DM avant ban : **{state}**")


@bot.command(name="setdeletedays")
@admin_only()
async def cmd_setdeletedays(ctx, days: int):
    """!setdeletedays <0-7>  —  Nombre de jours de messages supprimés lors du ban."""
    if not 0 <= days <= 7:
        await ctx.send("❌ La valeur doit être entre 0 et 7.")
        return
    config["delete_messages_days"] = days
    save_config(config)
    await ctx.send(f"✅ Suppression des messages des **{days}** derniers jour(s) lors d'un ban.")


# ════════════════════════════════════════════════════════════════════════════
#  COMMANDES — HISTORIQUE & STATUT
# ════════════════════════════════════════════════════════════════════════════

@bot.command(name="banhistory")
@admin_only()
async def cmd_banhistory(ctx, count: int = 10):
    """!banhistory [n]  —  Affiche les n derniers bans effectués par le bot (défaut 10)."""
    history = config["ban_history"]
    if not history:
        await ctx.send("Aucun ban enregistré.")
        return
    recent = history[-(min(count, 20)):][::-1]
    lines = []
    for entry in recent:
        ts = entry["timestamp"][:10]
        lines.append(f"`{ts}` **{entry['user']}** (`{entry['user_id']}`) — rôle : `{entry['role']}`")
    await ctx.send(f"**{len(recent)} dernier(s) ban(s) :**\n" + "\n".join(lines))


@bot.command(name="clearhistory")
@admin_only()
async def cmd_clearhistory(ctx):
    """!clearhistory  —  Vide l'historique des bans."""
    config["ban_history"] = []
    save_config(config)
    await ctx.send("✅ Historique des bans vidé.")


@bot.command(name="status")
@admin_only()
async def cmd_status(ctx):
    """!status  —  Affiche la configuration complète du bot."""
    log_mention = f"<#{config['log_channel_id']}>" if config["log_channel_id"] else "*(désactivé)*"
    roles = ", ".join(f"`{r}`" for r in config["banned_roles"]) or "*(aucun)*"

    embed = discord.Embed(title="⚙️ Configuration du bot", color=discord.Color.blurple())
    embed.add_field(name="Rôles interdits", value=roles, inline=False)
    embed.add_field(name="Salon de log", value=log_mention, inline=True)
    embed.add_field(name="DM avant ban", value="✅" if config["dm_user_before_ban"] else "❌", inline=True)
    embed.add_field(name="Suppression messages", value=f"{config['delete_messages_days']} jour(s)", inline=True)
    embed.add_field(name="Raison (audit log)", value=f"> {config['ban_reason']}", inline=False)
    embed.add_field(name="Message DM", value=f"> {config['dm_message']}", inline=False)
    embed.add_field(name="Bans enregistrés", value=str(len(config["ban_history"])), inline=True)
    await ctx.send(embed=embed)


@bot.command(name="bothelp")
async def cmd_help(ctx):
    """!bothelp  —  Liste toutes les commandes disponibles."""
    embed = discord.Embed(title="📖 Commandes du bot", color=discord.Color.blurple())
    embed.add_field(name="Rôles interdits", value=(
        "`!addrole <nom>`\n`!removerole <nom>`\n`!listroles`"
    ), inline=False)
    embed.add_field(name="Salon de log", value=(
        "`!setlog [#salon]`\n`!removelog`"
    ), inline=False)
    embed.add_field(name="Comportement du ban", value=(
        "`!setreason <texte>`\n`!setdm <message>`\n`!toggledm`\n`!setdeletedays <0-7>`"
    ), inline=False)
    embed.add_field(name="Historique & statut", value=(
        "`!status`\n`!banhistory [n]`\n`!clearhistory`\n`!bothelp`"
    ), inline=False)
    embed.set_footer(text="Toutes les commandes requièrent la permission Administrateur.")
    await ctx.send(embed=embed)


bot.run(TOKEN)
