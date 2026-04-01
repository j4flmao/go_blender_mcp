bl_info = {
    "name": "Blender MCP Bridge (Go)",
    "blender": (3, 6, 0),
    "version": (2, 0, 0),
    "category": "Development",
    "description": "TCP bridge for Go MCP server — lightweight, low context",
}

import bpy
import json
import os
import queue
import socket
import threading
import traceback
import urllib.request
from bpy.props import BoolProperty, StringProperty

HOST = "127.0.0.1"
PORT = 9876

_running = False
_server_thread = None
_req_queue = queue.Queue()
_timer_registered = False


def _prefs():
    addon = bpy.context.preferences.addons.get(__name__)
    if not addon:
        return None
    return addon.preferences


def _env_lines(prefs):
    if not prefs:
        return []

    lines = []
    if prefs.enable_polyhaven:
        lines.append('ENABLE_PH="1"')
    if prefs.enable_hyper3d:
        lines.append('ENABLE_H3D="1"')
        lines.append(f'HYPER3D_KEY="{prefs.hyper3d_key}"')
    if prefs.enable_sketchfab:
        lines.append('ENABLE_SF="1"')
        lines.append(f'SKETCHFAB_KEY="{prefs.sketchfab_key}"')
    if prefs.enable_hunyuan:
        lines.append('ENABLE_HY="1"')
        lines.append(f'HUNYUAN_KEY="{prefs.hunyuan_key}"')
    return lines

class MCP_Preferences(bpy.types.AddonPreferences):
    bl_idname = __name__

    enable_polyhaven: BoolProperty(name="Poly Haven", default=False)
    enable_hyper3d: BoolProperty(name="Hyper3D Rodin", default=False)
    hyper3d_key: StringProperty(name="HYPER3D_KEY", default="", subtype="PASSWORD")
    enable_sketchfab: BoolProperty(name="Sketchfab", default=False)
    sketchfab_key: StringProperty(name="SKETCHFAB_KEY", default="", subtype="PASSWORD")
    enable_hunyuan: BoolProperty(name="Hunyuan", default=False)
    hunyuan_key: StringProperty(name="HUNYUAN_KEY", default="", subtype="PASSWORD")

    def draw(self, context):
        layout = self.layout
        col = layout.column(align=True)
        col.prop(self, "enable_polyhaven")
        col.prop(self, "enable_hyper3d")
        if self.enable_hyper3d:
            col.prop(self, "hyper3d_key")
        col.prop(self, "enable_sketchfab")
        if self.enable_sketchfab:
            col.prop(self, "sketchfab_key")
        col.prop(self, "enable_hunyuan")
        if self.enable_hunyuan:
            col.prop(self, "hunyuan_key")


class MCP_OT_CopyEnv(bpy.types.Operator):
    bl_idname = "mcp.copy_env"
    bl_label = "Copy Claude Env"

    def execute(self, ctx):
        prefs = _prefs()
        lines = _env_lines(prefs)
        if not lines:
            self.report({"WARNING"}, "No integrations enabled")
            return {"CANCELLED"}
        bpy.context.window_manager.clipboard = "\n".join(lines)
        self.report({"INFO"}, "Copied env to clipboard")
        return {"FINISHED"}


def cmd_get_scene_info(p):
    s = bpy.context.scene
    return {
        "name": s.name,
        "frame": s.frame_current,
        "fps": s.render.fps,
        "objects": len(s.objects),
        "engine": s.render.engine,
    }


def cmd_list_objects(p):
    limit = int(p.get("limit", 50))
    objs = list(bpy.context.scene.objects)[:limit]
    return {"objects": [{"name": o.name, "type": o.type} for o in objs]}


def cmd_get_object(p):
    obj = bpy.data.objects.get(p["name"])
    if not obj:
        return {"error": f"Not found: {p['name']}"}
    return {
        "name": obj.name,
        "type": obj.type,
        "location": [round(v, 4) for v in obj.location],
        "rotation": [round(v, 4) for v in obj.rotation_euler],
        "scale": [round(v, 4) for v in obj.scale],
        "visible": obj.visible_get(),
    }


def cmd_create_object(p):
    t = p.get("type", "CUBE").upper()
    name = p.get("name", "Object")
    loc = p.get("location", [0, 0, 0])
    bpy.ops.object.select_all(action="DESELECT")
    dispatch = {
        "CUBE": lambda: bpy.ops.mesh.primitive_cube_add(location=loc),
        "SPHERE": lambda: bpy.ops.mesh.primitive_uv_sphere_add(location=loc),
        "CYLINDER": lambda: bpy.ops.mesh.primitive_cylinder_add(location=loc),
        "PLANE": lambda: bpy.ops.mesh.primitive_plane_add(location=loc),
        "CONE": lambda: bpy.ops.mesh.primitive_cone_add(location=loc),
        "TORUS": lambda: bpy.ops.mesh.primitive_torus_add(location=loc),
        "EMPTY": lambda: bpy.ops.object.empty_add(location=loc),
        "CAMERA": lambda: bpy.ops.object.camera_add(location=loc),
        "LIGHT": lambda: bpy.ops.object.light_add(location=loc),
    }
    fn = dispatch.get(t)
    if not fn:
        return {"error": f"Unknown type: {t}. Valid: {','.join(dispatch)}"}
    fn()
    obj = bpy.context.active_object
    obj.name = name
    return {"created": obj.name, "location": loc}


def cmd_move_object(p):
    obj = bpy.data.objects.get(p["name"])
    if not obj:
        return {"error": f"Not found: {p['name']}"}
    if "location" in p:
        obj.location = p["location"]
    if "rotation" in p:
        obj.rotation_euler = p["rotation"]
    if "scale" in p:
        obj.scale = p["scale"]
    return {"updated": obj.name, "location": [round(v, 4) for v in obj.location]}


def cmd_delete_object(p):
    obj = bpy.data.objects.get(p["name"])
    if not obj:
        return {"error": f"Not found: {p['name']}"}
    name = obj.name
    bpy.data.objects.remove(obj, do_unlink=True)
    return {"deleted": name}


def cmd_set_material(p):
    obj = bpy.data.objects.get(p["name"])
    if not obj:
        return {"error": f"Not found: {p['name']}"}
    color = p.get("color", [0.8, 0.8, 0.8, 1.0])
    metallic = p.get("metallic", 0.0)
    roughness = p.get("roughness", 0.5)
    mat_name = p.get("material_name", f"{obj.name}_mat")

    mat = bpy.data.materials.new(name=mat_name)
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes.get("Principled BSDF")
    if bsdf:
        bsdf.inputs["Base Color"].default_value = color
        bsdf.inputs["Metallic"].default_value = metallic
        bsdf.inputs["Roughness"].default_value = roughness

    if obj.data and hasattr(obj.data, "materials"):
        if obj.data.materials:
            obj.data.materials[0] = mat
        else:
            obj.data.materials.append(mat)
    return {"material": mat_name, "object": obj.name}


def cmd_render_scene(p):
    output = p.get("output_path", "/tmp/blender_render.png")
    width = int(p.get("width", 1920))
    height = int(p.get("height", 1080))
    samples = int(p.get("samples", 64))
    s = bpy.context.scene
    s.render.filepath = output
    s.render.resolution_x = width
    s.render.resolution_y = height
    if s.render.engine == "CYCLES" and hasattr(s, "cycles"):
        s.cycles.samples = samples
    bpy.ops.render.render(write_still=True)
    return {"rendered": output, "size": f"{width}x{height}"}


def cmd_set_render_engine(p):
    engine = p.get("engine", "CYCLES").upper()
    valid = {"CYCLES", "BLENDER_EEVEE", "BLENDER_WORKBENCH"}
    if engine not in valid:
        return {"error": f"Invalid. Valid: {','.join(valid)}"}
    bpy.context.scene.render.engine = engine
    return {"engine": engine}


def cmd_exec_python(p):
    code = p.get("code", "")
    if not code:
        return {"error": "No code provided"}
    try:
        ns = {}
        exec(code, {"bpy": bpy}, ns)
        return {"result": str(ns.get("result", "ok"))[:500]}
    except Exception:
        return {"error": traceback.format_exc(limit=3)}


def cmd_get_settings(p):
    prefs = _prefs()
    if not prefs:
        return {
            "enable_polyhaven": False,
            "enable_hyper3d": False,
            "enable_sketchfab": False,
            "enable_hunyuan": False,
            "hyper3d_key": "",
            "sketchfab_key": "",
            "hunyuan_key": "",
        }
    return {
        "enable_polyhaven": bool(prefs.enable_polyhaven),
        "enable_hyper3d": bool(prefs.enable_hyper3d),
        "enable_sketchfab": bool(prefs.enable_sketchfab),
        "enable_hunyuan": bool(prefs.enable_hunyuan),
        "hyper3d_key": str(prefs.hyper3d_key),
        "sketchfab_key": str(prefs.sketchfab_key),
        "hunyuan_key": str(prefs.hunyuan_key),
    }


def cmd_import_polyhaven_hdri(p):
    name = p.get("name", "")
    resolution = p.get("resolution", "1k")
    if not name:
        return {"error": "name required (e.g. 'autumn_field')"}
    url = f"https://dl.polyhaven.org/file/ph-assets/HDRIs/hdr/{resolution}/{name}_{resolution}.hdr"
    dest = f"/tmp/ph_{name}_{resolution}.hdr"
    try:
        urllib.request.urlretrieve(url, dest)
    except Exception as e:
        return {"error": f"Download failed: {e}"}

    world = bpy.context.scene.world or bpy.data.worlds.new("World")
    bpy.context.scene.world = world
    world.use_nodes = True
    nt = world.node_tree
    nt.nodes.clear()
    bg = nt.nodes.new("ShaderNodeBackground")
    env = nt.nodes.new("ShaderNodeTexEnvironment")
    out = nt.nodes.new("ShaderNodeOutputWorld")
    env.image = bpy.data.images.load(dest)
    nt.links.new(env.outputs["Color"], bg.inputs["Color"])
    nt.links.new(bg.outputs["Background"], out.inputs["Surface"])
    return {"hdri": name, "resolution": resolution, "path": dest}


def cmd_import_polyhaven_texture(p):
    name = p.get("name", "")
    resolution = p.get("resolution", "1k")
    tex_type = p.get("type", "diff")
    if not name:
        return {"error": "name required"}
    url = f"https://dl.polyhaven.org/file/ph-assets/Textures/jpg/{resolution}/{name}/{name}_{tex_type}_{resolution}.jpg"
    dest = f"/tmp/ph_{name}_{tex_type}_{resolution}.jpg"
    try:
        urllib.request.urlretrieve(url, dest)
    except Exception as e:
        return {"error": f"Download failed: {e}"}
    img = bpy.data.images.load(dest)
    return {"texture": name, "type": tex_type, "path": dest, "image": img.name}


def cmd_import_polyhaven_model(p):
    name = p.get("name", "")
    resolution = p.get("resolution", "1k")
    if not name:
        return {"error": "name required"}
    url = f"https://dl.polyhaven.org/file/ph-assets/Models/blend/{name}/{name}.blend"
    dest = f"/tmp/ph_{name}.blend"
    try:
        urllib.request.urlretrieve(url, dest)
    except Exception as e:
        return {"error": f"Download failed: {e}"}

    with bpy.data.libraries.load(dest) as (src, dst):
        dst.objects = src.objects[:]
    for obj in dst.objects:
        if obj:
            bpy.context.collection.objects.link(obj)
    return {"model": name, "objects": [o.name for o in dst.objects if o]}


COMMANDS = {
    "get_settings": cmd_get_settings,
    "get_scene_info": cmd_get_scene_info,
    "list_objects": cmd_list_objects,
    "get_object": cmd_get_object,
    "create_object": cmd_create_object,
    "move_object": cmd_move_object,
    "delete_object": cmd_delete_object,
    "set_material": cmd_set_material,
    "render_scene": cmd_render_scene,
    "set_render_engine": cmd_set_render_engine,
    "exec_python": cmd_exec_python,
    "import_polyhaven_hdri": cmd_import_polyhaven_hdri,
    "import_polyhaven_texture": cmd_import_polyhaven_texture,
    "import_polyhaven_model": cmd_import_polyhaven_model,
}


def handle(data):
    cmd = data.get("cmd")
    fn = COMMANDS.get(cmd)
    if not fn:
        return {"error": f"Unknown: {cmd}. Available: {','.join(COMMANDS)}"}
    try:
        return fn(data.get("params", {}))
    except Exception:
        return {"error": traceback.format_exc(limit=2)}


def _process_queue():
    if not _running:
        global _timer_registered
        _timer_registered = False
        return None

    while True:
        try:
            data, ev, holder = _req_queue.get_nowait()
        except queue.Empty:
            break
        try:
            holder["result"] = handle(data)
        except Exception:
            holder["result"] = {"error": traceback.format_exc(limit=2)}
        ev.set()

    return 0.05


def _ensure_timer():
    global _timer_registered
    if _timer_registered:
        return
    bpy.app.timers.register(_process_queue, first_interval=0.05, persistent=True)
    _timer_registered = True


def _drain_queue_on_stop():
    while True:
        try:
            _, ev, holder = _req_queue.get_nowait()
        except queue.Empty:
            break
        holder["result"] = {"error": "bridge stopped"}
        ev.set()


def client_handler(conn):
    with conn:
        try:
            raw = b""
            while True:
                chunk = conn.recv(4096)
                if not chunk:
                    break
                raw += chunk
                try:
                    json.loads(raw)
                    break
                except json.JSONDecodeError:
                    continue
            if not raw:
                return
            data = json.loads(raw)
            ev = threading.Event()
            holder = {}
            _req_queue.put((data, ev, holder))
            if not ev.wait(120.0):
                result = {"error": "timeout"}
            else:
                result = holder.get("result", {"error": "no result"})
        except Exception as e:
            result = {"error": str(e)}
        conn.sendall(json.dumps(result).encode())


def server_loop():
    global _running
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        s.bind((HOST, PORT))
        s.listen(5)
        s.settimeout(1.0)
        while _running:
            try:
                conn, _ = s.accept()
                threading.Thread(target=client_handler, args=(conn,), daemon=True).start()
            except socket.timeout:
                continue


class MCP_OT_Start(bpy.types.Operator):
    bl_idname = "mcp.start"
    bl_label = "Start Bridge"

    def execute(self, ctx):
        global _running, _server_thread
        if not _running:
            _running = True
            _ensure_timer()
            _server_thread = threading.Thread(target=server_loop, daemon=True)
            _server_thread.start()
            self.report({"INFO"}, f"Bridge on {HOST}:{PORT}")
        return {"FINISHED"}


class MCP_OT_Stop(bpy.types.Operator):
    bl_idname = "mcp.stop"
    bl_label = "Stop Bridge"

    def execute(self, ctx):
        global _running
        _running = False
        _drain_queue_on_stop()
        self.report({"INFO"}, "Bridge stopped")
        return {"FINISHED"}


class MCP_PT_Panel(bpy.types.Panel):
    bl_label = "Blender MCP Go"
    bl_idname = "MCP_PT_go_panel"
    bl_space_type = "VIEW_3D"
    bl_region_type = "UI"
    bl_category = "MCP"

    def draw(self, ctx):
        prefs = _prefs()
        col = self.layout.column(align=True)
        col.label(text=f"Port: {PORT}", icon="NETWORK_DRIVE")
        col.operator("mcp.start", icon="PLAY")
        col.operator("mcp.stop", icon="SNAP_FACE")
        col.separator()
        col.label(text="● Running" if _running else "○ Stopped", icon="CHECKMARK" if _running else "X")

        box = col.box()
        box.label(text="Integrations", icon="PREFERENCES")
        if not prefs:
            box.label(text="Addon preferences not found")
            return

        box.prop(prefs, "enable_polyhaven")

        box.prop(prefs, "enable_hyper3d")
        if prefs.enable_hyper3d:
            box.prop(prefs, "hyper3d_key", text="HYPER3D_KEY")

        box.prop(prefs, "enable_sketchfab")
        if prefs.enable_sketchfab:
            box.prop(prefs, "sketchfab_key", text="SKETCHFAB_KEY")

        box.prop(prefs, "enable_hunyuan")
        if prefs.enable_hunyuan:
            box.prop(prefs, "hunyuan_key", text="HUNYUAN_KEY")

        env = _env_lines(prefs)
        if env:
            box.separator()
            box.operator("mcp.copy_env", icon="COPYDOWN")


_classes = (MCP_Preferences, MCP_OT_CopyEnv, MCP_OT_Start, MCP_OT_Stop, MCP_PT_Panel)


def register():
    for cls in _classes:
        bpy.utils.register_class(cls)


def unregister():
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)


if __name__ == "__main__":
    register()

