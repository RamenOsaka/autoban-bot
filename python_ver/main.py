"""
Auto-Ban Bot — Slash commands via Cog
All commands grouped under /autoban
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


# ─── Cog ─────────────────────────────────────────────────────────────────────

class AutoBanCog(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot
        self.config = load_config()

    group = app_commands.Group(name="autoban", description="Auto-ban bot commands.")

    # ── Helpers ──────────────────────────────────────────────────────────────

    def is_banned_role(self, role: discord.Role) -> bool:
        for entry in self.config["banned_roles"]:
            if isinstance(entry, int) and role.id == entry:
                return True
            if isinstance(entry, str) and role.name.lower() == entry.lower():
                return True
        return False

    async def send_log(self, guild: discord.Guild, embed: discord.Embed):
        if self.config["log_channel_id"]:
            ch = guild.get_channel(int(self.config["log_channel_id"]))
            if ch:
                try:
                    await ch.send(embed=embed)
                except discord.Forbidden:
                    logging.warning("Cannot write to log channel.")

    async def do_ban(self, member: discord.Member, trigger_role_name: str):
        reason = f"{self.config['ban_reason']} (role: {trigger_role_name})"

        if self.config["dm_user_before_ban"]:
            try:
                await member.send(self.config["dm_message"])
            except discord.Forbidden:
                logging.info(f"Could not DM {member} (DMs closed).")

        try:
            await member.ban(
                reason=reason,
                delete_message_days=max(0, min(7, int(self.config["delete_messages_days"])))
            )
            timestamp = datetime.now(timezone.utc).isoformat()
            logging.info(f"Banned: {member} ({member.id}) — {reason}")

            self.config["ban_history"].append({
                "user": str(member),
                "user_id": member.id,
                "role": trigger_role_name,
                "reason": reason,
                "timestamp": timestamp
            })
            self.config["ban_history"] = self.config["ban_history"][-200:]
            save_config(self.config)

            embed = discord.Embed(
                title="🔨 Automatic Ban",
                color=discord.Color.red(),
                timestamp=datetime.now(timezone.utc)
            )
            embed.add_field(name="User", value=f"{member} (`{member.id}`)", inline=True)
            embed.add_field(name="Trigger role", value=trigger_role_name, inline=True)
            embed.add_field(name="Reason", value=reason, inline=False)
            embed.set_footer(text=f"DM sent: {'yes' if self.config['dm_user_before_ban'] else 'no'}")
            await self.send_log(member.guild, embed)

        except discord.Forbidden:
            logging.warning(f"Insufficient permissions to ban {member}.")
        except discord.HTTPException as e:
            logging.error(f"HTTP error while banning {member}: {e}")

    # ── Events ────────────────────────────────────────────────────────────────

    @commands.Cog.listener()
    async def on_member_update(self, before: discord.Member, after: discord.Member):
        added_roles = set(after.roles) - set(before.roles)
        for role in added_roles:
            if self.is_banned_role(role):
                await self.do_ban(after, role.name)
                break

    @commands.Cog.listener()
    async def on_member_join(self, member: discord.Member):
        for role in member.roles:
            if self.is_banned_role(role):
                await self.do_ban(member, role.name)
                break

    # ── Permission check ──────────────────────────────────────────────────────

    @staticmethod
    def admin_only(interaction: discord.Interaction) -> bool:
        return interaction.user.guild_permissions.administrator

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban addrole
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="addrole", description="Add a role to the auto-ban list.")
    @app_commands.describe(role="The role that will trigger an automatic ban.")
    async def cmd_addrole(self, interaction: discord.Interaction, role: discord.Role):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        if role.name in self.config["banned_roles"]:
            await interaction.response.send_message(f"⚠️ **{role.name}** is already in the ban list.", ephemeral=True)
            return
        self.config["banned_roles"].append(role.name)
        save_config(self.config)
        await interaction.response.send_message(f"✅ Role **{role.name}** added to the auto-ban list.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban removerole
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="removerole", description="Remove a role from the auto-ban list.")
    @app_commands.describe(role="The role to remove from the auto-ban list.")
    async def cmd_removerole(self, interaction: discord.Interaction, role: discord.Role):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        if role.name not in self.config["banned_roles"]:
            await interaction.response.send_message(f"❌ **{role.name}** is not in the ban list.", ephemeral=True)
            return
        self.config["banned_roles"].remove(role.name)
        save_config(self.config)
        await interaction.response.send_message(f"✅ Role **{role.name}** removed from the auto-ban list.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban listroles
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="listroles", description="List all roles currently on the auto-ban list.")
    async def cmd_listroles(self, interaction: discord.Interaction):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        if not self.config["banned_roles"]:
            await interaction.response.send_message("No forbidden roles configured.", ephemeral=True)
            return
        lines = "\n".join(f"• `{r}`" for r in self.config["banned_roles"])
        await interaction.response.send_message(
            f"**Forbidden roles ({len(self.config['banned_roles'])}):**\n{lines}", ephemeral=True
        )

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban setlog
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="setlog", description="Set the channel where bans are logged.")
    @app_commands.describe(channel="The text channel to send ban logs to.")
    async def cmd_setlog(self, interaction: discord.Interaction, channel: discord.TextChannel):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["log_channel_id"] = channel.id
        save_config(self.config)
        await interaction.response.send_message(f"✅ Log channel set to {channel.mention}.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban removelog
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="removelog", description="Disable ban logging.")
    async def cmd_removelog(self, interaction: discord.Interaction):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["log_channel_id"] = None
        save_config(self.config)
        await interaction.response.send_message("✅ Log channel disabled.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban setreason
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="setreason", description="Set the ban reason shown in the Discord audit log.")
    @app_commands.describe(reason="The reason text shown in the audit log.")
    async def cmd_setreason(self, interaction: discord.Interaction, reason: str):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["ban_reason"] = reason
        save_config(self.config)
        await interaction.response.send_message(f"✅ Ban reason updated:\n> {reason}")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban setdm
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="setdm", description="Set the DM message sent to the user before they are banned.")
    @app_commands.describe(message="The message the user will receive before being banned.")
    async def cmd_setdm(self, interaction: discord.Interaction, message: str):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["dm_message"] = message
        save_config(self.config)
        await interaction.response.send_message(f"✅ DM message updated:\n> {message}")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban toggledm
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="toggledm", description="Enable or disable the DM sent to users before being banned.")
    async def cmd_toggledm(self, interaction: discord.Interaction):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["dm_user_before_ban"] = not self.config["dm_user_before_ban"]
        save_config(self.config)
        state = "enabled ✅" if self.config["dm_user_before_ban"] else "disabled ❌"
        await interaction.response.send_message(f"DM before ban: **{state}**")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban setdeletedays
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="setdeletedays", description="Set how many days of messages to delete when banning (0–7).")
    @app_commands.describe(days="Number of days of messages to delete (0 = none, 7 = max).")
    async def cmd_setdeletedays(self, interaction: discord.Interaction, days: int):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        if not 0 <= days <= 7:
            await interaction.response.send_message("❌ Value must be between 0 and 7.", ephemeral=True)
            return
        self.config["delete_messages_days"] = days
        save_config(self.config)
        await interaction.response.send_message(f"✅ Will delete the last **{days}** day(s) of messages on ban.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban banhistory
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="banhistory", description="Show the most recent bans performed by the bot.")
    @app_commands.describe(count="Number of recent bans to show (max 20).")
    async def cmd_banhistory(self, interaction: discord.Interaction, count: int = 10):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        history = self.config["ban_history"]
        if not history:
            await interaction.response.send_message("No bans recorded.", ephemeral=True)
            return
        recent = history[-(min(count, 20)):][::-1]
        lines = [
            f"`{e['timestamp'][:10]}` **{e['user']}** (`{e['user_id']}`) — role: `{e['role']}`"
            for e in recent
        ]
        await interaction.response.send_message(
            f"**Last {len(recent)} ban(s):**\n" + "\n".join(lines), ephemeral=True
        )

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban clearhistory
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="clearhistory", description="Clear the bot's ban history log.")
    async def cmd_clearhistory(self, interaction: discord.Interaction):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        self.config["ban_history"] = []
        save_config(self.config)
        await interaction.response.send_message("✅ Ban history cleared.")

    # ════════════════════════════════════════════════════════════════════════
    #  /autoban status
    # ════════════════════════════════════════════════════════════════════════

    @group.command(name="status", description="Display the bot's current configuration.")
    async def cmd_status(self, interaction: discord.Interaction):
        if not self.admin_only(interaction):
            await interaction.response.send_message("❌ Administrator permission required.", ephemeral=True)
            return
        log_mention = f"<#{self.config['log_channel_id']}>" if self.config["log_channel_id"] else "*(disabled)*"
        roles = ", ".join(f"`{r}`" for r in self.config["banned_roles"]) or "*(none)*"

        embed = discord.Embed(title="⚙️ Bot Configuration", color=discord.Color.blurple())
        embed.add_field(name="Forbidden roles", value=roles, inline=False)
        embed.add_field(name="Log channel", value=log_mention, inline=True)
        embed.add_field(name="DM before ban", value="✅" if self.config["dm_user_before_ban"] else "❌", inline=True)
        embed.add_field(name="Delete messages", value=f"{self.config['delete_messages_days']} day(s)", inline=True)
        embed.add_field(name="Audit log reason", value=f"> {self.config['ban_reason']}", inline=False)
        embed.add_field(name="DM message", value=f"> {self.config['dm_message']}", inline=False)
        embed.add_field(name="Bans recorded", value=str(len(self.config["ban_history"])), inline=True)
        await interaction.response.send_message(embed=embed, ephemeral=True)


# ─── Bot setup ───────────────────────────────────────────────────────────────

intents = discord.Intents.default()
intents.members = True
intents.message_content = True

bot = commands.Bot(command_prefix="!", intents=intents)


@bot.event
async def on_ready():
    await bot.add_cog(AutoBanCog(bot))
    await bot.tree.sync()
    logging.info(f"✅ Logged in as {bot.user} ({bot.user.id})")
    logging.info(f"Slash commands synced under /autoban")


bot.run(TOKEN)
