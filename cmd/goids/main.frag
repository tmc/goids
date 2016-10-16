#version 330
//uniform sampler2D tex;
//in vec2 fragTexCoord;
uniform vec4 color;
out vec4 outputColor;
void main() {
    outputColor = color;
}
