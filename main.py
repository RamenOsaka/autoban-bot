"""
Auto-Ban Bot — Slash commands version
Bans automatically any member who receives a forbidden role.
All configuration is managed via Discord slash commands and persisted in config.json.
"""

import discord
from discord import app_commands
from discord.ext import commands
import json
import os
import logging
from datetime import datetime, timezone

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("bot.log", encoding="utf-8"),
        logging.StreamHandler()
    ]
)

TOKEN = os.environ.get("DISCORD_TOKEN", "YOUR_TOKEN_HERE")
CONFIG_FILE = "config.json"

DEFAULT_CONFIG = {
    "banned_roles": [],
    "log_channel_id": None,
    "ban_reason": "Automatic ban: forbidden role detected.",
    "delete_messages_days": 1,
    "dm_user_before_ban": True,
    "dm_message": "You have been automatically banned from the server for receiving a forbidden role.",
    "ban_history": []
}


def load_config() -> dict:
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
        for key, value in DEFAULT_CONFIG.items():
            data.setdefault(key, value)
        return data
    return dict(DEFAULT_CONFIG)


def save_config(cfg: dict):
    with open(CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(cfg, f, indent=2, ensure_ascii=False)


config = load_config()

intents = discord.Intents.default()
intents.members = True
intents.message_content = True

bot = commands.Bot(command_prefix="!", intents=intents)
tree = bot.tree


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
                logging.warning("Cannot write to log channel.")


async def ban_member(member: discord.Member, trigger_role_name: str):
    reason = f"{config['ban_reason']} (role: {trigger_role_name})"

    if config["dm_user_before_ban"]:
        try:
            await member.send(config["dm_message"])
        except discord.Forbidden:
            logging.info(f"Could not DM {member} (DMs closed).")

    try:
        await member.ban(
            reason=reason,
            delete_message_days=max(0, min(7, int(config["delete_messages_days"])))
        )
        timestamp = datetime.now(timezone.utc).isoformat()
        logging.info(f"Banned: {member} ({member.id}) — {reason}")

        config["ban_history"].append({
            "user": str(member),
            "user_id": member.id,
            "role": trigger_role_name,
            "reason": reason,
            "timestamp": timestamp
        })
        config["ban_history"] = config["ban_history"][-200:]
        save_config(config)

        embed = discord.Embed(
            title="🔨 Automatic Ban",
            color=discord.Color.red(),
            timestamp=datetime.now(timezone.utc)
        )
        embed.add_field(name="User", value=f"{member} (`{member.id}`)", inline=True)
        embed.add_field(name="Trigger role", value=trigger_role_name, inline=True)
        embed.add_field(name="Reason", value=reason, inline=False)
        embed.set_footer(text=f"DM sent: {'yes' if config['dm_user_before_ban'] else 'no'}")
        await send_log(member.guild, embed)

    except discord.Forbidden:
        logging.warning(f"Insufficient permissions to ban {member}.")
    except discord.HTTPException as e:
        logging.error(f"HTTP error while banning {member}: {e}")


# ─── Events ──────────────────────────────────────────────────────────────────

@bot.event
async def on_ready():
    await tree.sync()
    logging.info(f"✅ Logged in as {bot.user} ({bot.user.id})")
    logging.info(f"Slash commands synced.")
    logging.info(f"Watched roles: {config['banned_roles']}")


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


# ─── Permission check ────────────────────────────────────────────────────────

def is_admin(interaction: discord.Interaction) -> bool:
    return interaction.user.guild_permissions.administrator


# ════════════════════════════════════════════════════════════════════════════
#  SLASH COMMANDS — BANNED ROLES
# ════════════════════════════════════════════════════════════════════════════

@tree.command(name="addrole", description="Add a role to the auto-ban list.")
@app_commands.describe(role="The role that will trigger an automatic ban.")
async def cmd_addrole(interaction: discord.Interaction, role: discord.Role):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    if role.name in config["banned_roles"] or role.id in config["banned_roles"]:
        await interaction.response.send_message(f"⚠️ **{role.name}** is already in the ban list.", ephemeral=True)
        return
    config["banned_roles"].append(role.name)
    save_config(config)
    await interaction.response.send_message(f"✅ Role **{role.name}** added to the auto-ban list.")


@tree.command(name="removerole", description="Remove a role from the auto-ban list.")
@app_commands.describe(role="The role to remove from the auto-ban list.")
async def cmd_removerole(interaction: discord.Interaction, role: discord.Role):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    if role.name not in config["banned_roles"]:
        await interaction.response.send_message(f"❌ **{role.name}** is not in the ban list.", ephemeral=True)
        return
    config["banned_roles"].remove(role.name)
    save_config(config)
    await interaction.response.send_message(f"✅ Role **{role.name}** removed from the auto-ban list.")


@tree.command(name="listroles", description="List all roles currently on the auto-ban list.")
async def cmd_listroles(interaction: discord.Interaction):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    if not config["banned_roles"]:
        await interaction.response.send_message("No forbidden roles configured.", ephemeral=True)
        return
    lines = "\n".join(f"• `{r}`" for r in config["banned_roles"])
    await interaction.response.send_message(
        f"**Forbidden roles ({len(config['banned_roles'])}):**\n{lines}", ephemeral=True
    )


# ════════════════════════════════════════════════════════════════════════════
#  SLASH COMMANDS — LOG CHANNEL
# ════════════════════════════════════════════════════════════════════════════

@tree.command(name="setlog", description="Set the channel where bans are logged.")
@app_commands.describe(channel="The text channel to send ban logs to.")
async def cmd_setlog(interaction: discord.Interaction, channel: discord.TextChannel):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["log_channel_id"] = channel.id
    save_config(config)
    await interaction.response.send_message(f"✅ Log channel set to {channel.mention}.")


@tree.command(name="removelog", description="Disable ban logging.")
async def cmd_removelog(interaction: discord.Interaction):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["log_channel_id"] = None
    save_config(config)
    await interaction.response.send_message("✅ Log channel disabled.")


# ════════════════════════════════════════════════════════════════════════════
#  SLASH COMMANDS — BAN BEHAVIOR
# ════════════════════════════════════════════════════════════════════════════

@tree.command(name="setreason", description="Set the ban reason shown in the Discord audit log.")
@app_commands.describe(reason="The reason text shown in the audit log.")
async def cmd_setreason(interaction: discord.Interaction, reason: str):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["ban_reason"] = reason
    save_config(config)
    await interaction.response.send_message(f"✅ Ban reason updated:\n> {reason}")


@tree.command(name="setdm", description="Set the DM message sent to the user before they are banned.")
@app_commands.describe(message="The message the user will receive before being banned.")
async def cmd_setdm(interaction: discord.Interaction, message: str):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["dm_message"] = message
    save_config(config)
    await interaction.response.send_message(f"✅ DM message updated:\n> {message}")


@tree.command(name="toggledm", description="Enable or disable the DM sent to users before being banned.")
async def cmd_toggledm(interaction: discord.Interaction):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["dm_user_before_ban"] = not config["dm_user_before_ban"]
    save_config(config)
    state = "enabled ✅" if config["dm_user_before_ban"] else "disabled ❌"
    await interaction.response.send_message(f"DM before ban: **{state}**")


@tree.command(name="setdeletedays", description="Set how many days of messages to delete when banning (0–7).")
@app_commands.describe(days="Number of days of messages to delete (0 = none, 7 = max).")
async def cmd_setdeletedays(interaction: discord.Interaction, days: int):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    if not 0 <= days <= 7:
        await interaction.response.send_message("❌ Value must be between 0 and 7.", ephemeral=True)
        return
    config["delete_messages_days"] = days
    save_config(config)
    await interaction.response.send_message(f"✅ Will delete the last **{days}** day(s) of messages on ban.")


# ════════════════════════════════════════════════════════════════════════════
#  SLASH COMMANDS — HISTORY & STATUS
# ════════════════════════════════════════════════════════════════════════════

@tree.command(name="banhistory", description="Show the most recent bans performed by the bot.")
@app_commands.describe(count="Number of recent bans to show (max 20).")
async def cmd_banhistory(interaction: discord.Interaction, count: int = 10):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    history = config["ban_history"]
    if not history:
        await interaction.response.send_message("No bans recorded.", ephemeral=True)
        return
    recent = history[-(min(count, 20)):][::-1]
    lines = []
    for entry in recent:
        ts = entry["timestamp"][:10]
        lines.append(f"`{ts}` **{entry['user']}** (`{entry['user_id']}`) — role: `{entry['role']}`")
    await interaction.response.send_message(
        f"**Last {len(recent)} ban(s):**\n" + "\n".join(lines), ephemeral=True
    )


@tree.command(name="clearhistory", description="Clear the bot's ban history log.")
async def cmd_clearhistory(interaction: discord.Interaction):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    config["ban_history"] = []
    save_config(config)
    await interaction.response.send_message("✅ Ban history cleared.")


@tree.command(name="status", description="Display the bot's current configuration.")
async def cmd_status(interaction: discord.Interaction):
    if not is_admin(interaction):
        await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
        return
    log_mention = f"<#{config['log_channel_id']}>" if config["log_channel_id"] else "*(disabled)*"
    roles = ", ".join(f"`{r}`" for r in config["banned_roles"]) or "*(none)*"

    embed = discord.Embed(title="⚙️ Bot Configuration", color=discord.Color.blurple())
    embed.add_field(name="Forbidden roles", value=roles, inline=False)
    embed.add_field(name="Log channel", value=log_mention, inline=True)
    embed.add_field(name="DM before ban", value="✅" if config["dm_user_before_ban"] else "❌", inline=True)
    embed.add_field(name="Delete messages", value=f"{config['delete_messages_days']} day(s)", inline=True)
    embed.add_field(name="Audit log reason", value=f"> {config['ban_reason']}", inline=False)
    embed.add_field(name="DM message", value=f"> {config['dm_message']}", inline=False)
    embed.add_field(name="Bans recorded", value=str(len(config["ban_history"])), inline=True)
    await interaction.response.send_message(embed=embed, ephemeral=True)


bot.run(TOKEN)
