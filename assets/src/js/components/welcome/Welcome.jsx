import Window from "../ui/Window";

export default function Welcome({ user }) {
  return (
    <div className="h-full flex items-center justify-center p-4">
      <Window title="Agent64" className="w-full max-w-[380px]">
        <p>welcome, <b>{user.username}</b></p>
      </Window>
    </div>
  );
}