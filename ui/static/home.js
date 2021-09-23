document.addEventListener("DOMContentLoaded", () => {
	const joinForm = document.getElementById("join");
	joinForm.addEventListener("submit", (e) => {
		const code = joinForm.querySelector("input[name=\"code\"]").value;
		joinForm.action = "/play/" + encodeURIComponent(code);
	});
	document.body.addEventListener("click", () => {
		for (const help of document.querySelectorAll(".message.show")) {
			help.classList.remove("show");
		}
	});
	const helps = document.getElementsByClassName("help");
	for (const help of helps) {
		help.addEventListener("click", (e) => {
			e.stopPropagation();
			console.log(e.target);
			e.target.parentElement.querySelector(".message").classList.toggle("show");
		})
	}
});
