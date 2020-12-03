document.addEventListener("DOMContentLoaded", () => {
	const joinForm = document.getElementById("join");
	joinForm.addEventListener("submit", (e) => {
		const code = joinForm.querySelector("input[name=\"code\"]").value;
		joinForm.action = "/play/" + encodeURIComponent(code);
	});
});
