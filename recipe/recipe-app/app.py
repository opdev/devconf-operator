from flask import Flask, request, jsonify, render_template, redirect, url_for
from models import create_connection
from datetime import datetime
import pymysql
import os
import psutil

app = Flask(__name__)

@app.route('/')
def index():
    with create_connection() as conn:
        cursor = conn.cursor()
        cursor.execute('SELECT * FROM recipes')
        recipes = cursor.fetchall()  # This should return a list of dictionaries
    return render_template('index.html', recipes=recipes)

@app.route('/recipes/<int:recipe_id>', methods=['GET', 'POST'])
def get_recipe_by_id(recipe_id):
    if request.method == 'POST':
        title = request.form['title']
        ingredients = request.form['ingredients']
        instructions = request.form['instructions']
        with create_connection() as conn:
            cursor = conn.cursor()
            cursor.execute('''
                UPDATE recipes
                SET title = %s, ingredients = %s, instructions = %s, updated_at = %s
                WHERE id = %s
            ''', (title, ingredients, instructions, datetime.now(), recipe_id))
            conn.commit()
        return redirect(url_for('index'))
    else:
        with create_connection() as conn:
            cursor = conn.cursor()
            cursor.execute('SELECT * FROM recipes WHERE id = %s', (recipe_id,))
            recipe = cursor.fetchone()
        return render_template('edit_recipe.html', recipe=recipe)

@app.route('/recipes/add', methods=['GET', 'POST'])
def add_recipe():
    if request.method == 'POST':
        title = request.form['title']
        ingredients = request.form['ingredients']
        instructions = request.form['instructions']
        with create_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute('''
                    INSERT INTO recipes (title, ingredients, instructions, created_at, updated_at)
                    VALUES (%s, %s, %s, %s, %s)
                ''', (title, ingredients, instructions, datetime.now(), datetime.now()))
                conn.commit()
        return redirect(url_for('index'))
    return render_template('add_recipe.html')

@app.route('/recipes/delete/<int:recipe_id>')
def delete_recipe(recipe_id):
    with create_connection() as conn:
        cursor = conn.cursor()
        cursor.execute('DELETE FROM recipes WHERE id = %s', (recipe_id,))
        conn.commit()
    return redirect(url_for('index'))

@app.route('/liveness', methods=['GET'])
def liveness():
    if check_application_health():
        return jsonify({"status": "ok"}), 200
    else:
        return jsonify({"status": "fail"}), 500

@app.route('/readiness', methods=['GET'])
def readiness():
    if check_database_connection():
        return jsonify({"status": "ok"}), 200
    else:
        return jsonify({"status": "fail"}), 500


if __name__ == '__main__':
    app.run(host='0.0.0.0')
    
# Example: Database connection status (for readiness check)
def check_database_connection():
    try:
        connection = pymysql.connect(
            host=os.environ.get('DB_HOST', 'localhost'),
            user=os.environ.get('DB_USER', 'root'),
            password=os.environ.get('DB_PASSWORD', ''),
            database=os.environ.get('DB_NAME', 'recipes'),
            port=int(os.environ.get('DB_PORT', 3306)),
            connect_timeout=5
        )
        connection.close()
        return True
    except pymysql.MySQLError as e:
        print(f"Database connection failed: {e}")
        return False

# Simplified application health check for liveness
def check_application_health():
    return True
