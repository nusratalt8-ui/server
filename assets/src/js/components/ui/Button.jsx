export const BTN_COLOR = {
  red:    { bg: "rgba(220,50,50,0.15)",  border: "rgba(220,50,50,0.4)",  text: "#e87070" },
  orange: { bg: "rgba(220,130,40,0.15)", border: "rgba(220,130,40,0.4)", text: "#e8a850" },
  green:  { bg: "rgba(50,180,80,0.15)",  border: "rgba(50,180,80,0.4)",  text: "#60c878" },
  blue:   { bg: "rgba(60,130,220,0.15)", border: "rgba(60,130,220,0.4)", text: "#70a8e8" },
};

export default function Button({ children, disabled, color, active, className = "", style, ...props }) {
  const c = color ? BTN_COLOR[color] || BTN_COLOR.blue : null;
  const tint = c
    ? { background: active ? c.border : c.bg, ...(disabled ? {} : { color: c.text }) }
    : null;
  return (
    <button disabled={disabled} className={"y2k-btn font-bold " + className} style={{ ...tint, ...style }} {...props}>
      {children}
    </button>
  );
}